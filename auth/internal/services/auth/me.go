package authservice

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"auth/internal/authctx"
	"auth/internal/events"
	"auth/internal/models/records"
	apperrors "auth/pkg/errors"
	"auth/pkg/utils"
)

const (
	authEventAccountDeleteStarted  = "account_delete_started"
	authEventAccountDeleteVerified = "account_delete_verified"
)

// GetMe - сервис получения текущей учетной записи
/*
	1. Получить identity_id из auth-контекста.
	2. Найти identity по id.
	3. Проверить, что аккаунт не удален и не заблокирован.
	4. Смаппить record в сервисный Identity DTO.
	5. Вернуть данные текущей учетной записи.
*/
func (s *AuthService) GetMe(ctx context.Context) (*Identity, error) {
	const op = "AuthService.GetMe"

	// 1. Получаем identity_id из auth-контекста.
	identityID, _, err := authctx.FromContext(ctx)
	if err != nil {
		return nil, apperrors.New(op, err)
	}

	// 2. Находим identity по id.
	identity, err := s.identities.GetByID(ctx, identityID)
	if err != nil {
		return nil, apperrors.New(op, err)
	}

	// 3. Проверяем, что аккаунт не удален и не заблокирован.
	if identity.Status == "deleted" {
		return nil, apperrors.New(op, ErrIdentityDeleted)
	}
	if identity.Status != identityStatusActive {
		return nil, apperrors.New(op, ErrIdentityNotActive)
	}

	// 4. Смаппим record в сервисный Identity DTO.
	// 5. Возвращаем данные текущей учетной записи.
	return &Identity{
		ID:            identity.ID.String(),
		Email:         identity.Email,
		EmailVerified: identity.EmailVerified,
		Status:        identity.Status,
		CreatedAt:     identity.CreatedAt,
	}, nil
}

// DeleteAccount - сервис запуска удаления аккаунта
/*
	1. Получить identity_id из auth-контекста.
	2. Найти identity и credentials по identity_id.
	3. Проверить переданный password через password_hash.
	4. Проверить, что нет активного account_delete_request или переиспользовать текущий pending request.
	5. Проверить rate limit отправки кода удаления аккаунта.
	6. Сгенерировать verification code purpose=account_delete.
	7. В транзакции создать account_delete_request, сохранить verification code и записать auth_event.
	8. После commit опубликовать Kafka-событие notification.email.account_delete_code.send (payload: identity_id, email, code, purpose).
	9. Отправка кода выполняется notification-service по Kafka-событию.
	10. Вернуть сообщение об отправке кода.
*/
func (s *AuthService) DeleteAccount(ctx context.Context, input *DeleteAccountInput) (string, error) {
	const op = "AuthService.DeleteAccount"

	// 1. Проверяем, что сервис может работать с транзакциями и rate limiter'ом.
	if s.transactions == nil {
		return "", apperrors.New(op, ErrTransactionsNotConfigured)
	}
	if s.rateLimiter == nil {
		return "", apperrors.New(op, ErrRateLimiterNotConfigured)
	}

	// 2. Получаем identity_id из auth-контекста.
	identityID, _, err := authctx.FromContext(ctx)
	if err != nil {
		return "", apperrors.New(op, err)
	}

	// 3. Находим identity и проверяем, что аккаунт active и email подтвержден.
	identity, err := s.identities.GetByID(ctx, identityID)
	if err != nil {
		return "", apperrors.New(op, err)
	}
	if identity.Status != identityStatusActive {
		return "", apperrors.New(op, ErrIdentityNotActive)
	}
	if !identity.EmailVerified {
		return "", apperrors.New(op, ErrEmailNotVerified)
	}

	// 4. Получаем credentials и проверяем password через password_hash.
	credential, err := s.credentials.GetByIdentityID(ctx, identityID)
	if err != nil {
		return "", apperrors.New(op, err)
	}
	if !utils.VerifyPassword(input.Password, credential.PasswordHash) {
		return "", apperrors.New(op, ErrCurrentPasswordIncorrect)
	}

	// 5. Проверяем, что нет активного account_delete_request или переиспользуем текущий pending request.
	existingRequest, err := s.accountDeleteRequests.GetActiveByIdentityID(ctx, identityID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return "", apperrors.New(op, err)
	}

	// 6. Проверяем rate limit отправки кода удаления аккаунта.
	if err := s.rateLimiter.Allow(ctx, "rate:verification_code:account_delete:"+identityID.String(), s.features.CodeResendCooldown, s.features.CodeResendWindow, int64(s.features.CodeResendMaxRequests)); err != nil {
		return "", apperrors.New(op, err)
	}

	// 7. Генерируем verification code purpose=account_delete.
	code, err := utils.GenerateVerificationCode()
	if err != nil {
		return "", apperrors.New(op, err)
	}

	// 8. В транзакции создаем account_delete_request (если нет существующего), сохраняем verification code и записываем auth_event.
	now := time.Now().UTC()
	expiresAt := now.Add(s.features.CodeTTL)
	err = s.transactions(ctx, func(repos Repositories) error {
		if existingRequest == nil {
			if err := repos.AccountDeleteRequests.Create(ctx, &records.AccountDeleteRequest{
				ID:         uuid.New(),
				IdentityID: identityID,
				Status:     accountDeleteStatusPending,
				ExpiresAt:  expiresAt,
				CreatedAt:  now,
				UpdatedAt:  now,
			}); err != nil {
				return err
			}
		}

		if err := repos.VerificationCodes.Create(ctx, &records.VerificationCode{
			ID:            uuid.New(),
			IdentityID:    &identityID,
			Email:         &identity.Email,
			Purpose:       verificationPurposeAccountDelete,
			CodeHash:      utils.HashSHA256(code),
			AttemptsCount: 0,
			MaxAttempts:   s.features.CodeMaxAttempts,
			ExpiresAt:     now.Add(s.features.CodeTTL),
			CreatedAt:     now,
		}); err != nil {
			return err
		}

		return repos.AuthEvents.Create(ctx, &records.AuthEvent{
			ID:         uuid.New(),
			IdentityID: &identityID,
			EventType:  authEventAccountDeleteStarted,
			CreatedAt:  now,
		})
	})
	if err != nil {
		return "", apperrors.New(op, err)
	}

	// 9. После commit публикуем Kafka-событие notification.email.account_delete_code.send.
	identityIDString := identityID.String()
	if err := s.events.PublishAccountDeleteCodeSend(ctx, events.VerificationCodeSendPayload{
		IdentityID: &identityIDString,
		Email:      identity.Email,
		Code:       code,
		Purpose:    verificationPurposeAccountDelete,
	}); err != nil {
		return "", apperrors.New(op, err)
	}

	// 10. Возвращаем сообщение об отправке кода.
	return accountDeleteCodeSentMessage, nil
}

// DeleteAccountVerify - сервис подтверждения удаления аккаунта
/*
	1. Получить identity_id из auth-контекста.
	2. Найти активный verification code purpose=account_delete по identity_id.
	3. Проверить, что attempts_count меньше max_attempts.
	4. Сравнить переданный code с сохраненным code_hash.
	5. Если код неверный, увеличить attempts_count и вернуть ошибку.
	6. Если код верный, найти активный account_delete_request со статусом pending.
	7. В транзакции пометить request verified, verification code consumed, soft-delete identity и отозвать все refresh-сессии.
	8. В той же транзакции записать auth_event удаления аккаунта.
	9. После commit опубликовать Kafka-событие identity.deleted (payload: identity_id).
	10. После commit опубликовать Kafka-событие profile.delete (payload: identity_id).
*/
func (s *AuthService) DeleteAccountVerify(ctx context.Context, input *DeleteAccountVerifyInput) error {
	const op = "AuthService.DeleteAccountVerify"

	// 1. Проверяем, что сервис может работать с транзакциями.
	if s.transactions == nil {
		return apperrors.New(op, ErrTransactionsNotConfigured)
	}

	// 2. Получаем identity_id из auth-контекста.
	identityID, _, err := authctx.FromContext(ctx)
	if err != nil {
		return apperrors.New(op, err)
	}

	// 3. Найти активный verification code purpose=account_delete по identity_id.
	code, err := s.verificationCodes.GetActiveByIdentityIDAndPurpose(ctx, identityID, verificationPurposeAccountDelete)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.New(op, ErrInvalidCode)
		}
		return apperrors.New(op, err)
	}

	// 4. Проверить, что attempts_count меньше max_attempts.
	if code.AttemptsCount >= code.MaxAttempts {
		return apperrors.New(op, ErrTooManyAttempts)
	}

	// 5. Сравнить переданный code с сохраненным code_hash.
	if utils.HashSHA256(input.Code) != code.CodeHash {
		// 5.1 Если код неверный, увеличить attempts_count и вернуть ошибку.
		if err := s.verificationCodes.IncrementAttempts(ctx, code.ID); err != nil {
			return apperrors.New(op, err)
		}
		return apperrors.New(op, ErrInvalidCode)
	}

	// 6. Если код верный, находим активный account_delete_request со статусом pending.
	request, err := s.accountDeleteRequests.GetActiveByIdentityID(ctx, identityID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.New(op, ErrAccountDeleteNotFound)
		}
		return apperrors.New(op, err)
	}

	// 7. В транзакции помечаем request verified, verification code consumed, soft-delete identity и отозваем все refresh-сессии.
	now := time.Now().UTC()
	err = s.transactions(ctx, func(repos Repositories) error {
		if err := repos.AccountDeleteRequests.SetStatus(ctx, request.ID, accountDeleteStatusVerified); err != nil {
			return err
		}
		if err := repos.VerificationCodes.Consume(ctx, code.ID); err != nil {
			return err
		}
		if err := repos.Identities.SoftDelete(ctx, identityID); err != nil {
			return err
		}
		if err := repos.Sessions.RevokeAllByIdentityID(ctx, identityID); err != nil {
			return err
		}
		// 8. В той же транзакции записываем auth_event удаления аккаунта.
		return repos.AuthEvents.Create(ctx, &records.AuthEvent{
			ID:         uuid.New(),
			IdentityID: &identityID,
			EventType:  authEventAccountDeleteVerified,
			CreatedAt:  now,
		})
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.New(op, ErrInvalidCode)
		}
		return apperrors.New(op, err)
	}

	// 9. После commit публикуем Kafka-событие identity.deleted.
	if err := s.events.PublishIdentityDeleted(ctx, events.IdentityDeletedPayload{
		IdentityID: identityID.String(),
	}); err != nil {
		return apperrors.New(op, err)
	}

	return nil
}

// DeleteAccountCodeResend - сервис повторной отправки кода удаления аккаунта
/*
	1. Получить identity_id из auth-контекста.
	2. Найти identity и проверить, что аккаунт active и email_verified.
	3. Найти активный account_delete_request со статусом pending.
	4. Проверить rate limit повторной отправки.
	5. Сгенерировать новый verification code purpose=account_delete.
	6. В транзакции сохранить verification code и записать auth_event.
	7. После commit опубликовать Kafka-событие notification.email.account_delete_code.send (payload: identity_id, email, code, purpose).
	8. Вернуть сообщение об отправке.
*/
func (s *AuthService) DeleteAccountCodeResend(ctx context.Context) (string, error) {
	const op = "AuthService.DeleteAccountCodeResend"

	// 1. Проверяем, что сервис может работать с транзакциями и rate limiter'ом.
	if s.transactions == nil {
		return "", apperrors.New(op, ErrTransactionsNotConfigured)
	}
	if s.rateLimiter == nil {
		return "", apperrors.New(op, ErrRateLimiterNotConfigured)
	}

	// 2. Получаем identity_id из auth-контекста.
	identityID, _, err := authctx.FromContext(ctx)
	if err != nil {
		return "", apperrors.New(op, err)
	}

	// 3. Находим identity и проверяем, что аккаунт active и email подтвержден.
	identity, err := s.identities.GetByID(ctx, identityID)
	if err != nil {
		return "", apperrors.New(op, err)
	}
	if identity.Status != identityStatusActive {
		return "", apperrors.New(op, ErrIdentityNotActive)
	}
	if !identity.EmailVerified {
		return "", apperrors.New(op, ErrEmailNotVerified)
	}

	// 4. Проверяем, что существует активный account_delete_request со статусом pending.
	if _, err := s.accountDeleteRequests.GetActiveByIdentityID(ctx, identityID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", apperrors.New(op, ErrAccountDeleteNotFound)
		}
		return "", apperrors.New(op, err)
	}

	// 5. Проверяем rate limit повторной отправки.
	if err := s.rateLimiter.Allow(ctx, "rate:verification_code:account_delete:"+identityID.String(), s.features.CodeResendCooldown, s.features.CodeResendWindow, int64(s.features.CodeResendMaxRequests)); err != nil {
		return "", apperrors.New(op, err)
	}

	// 6. Генерируем новый verification code purpose=account_delete.
	code, err := utils.GenerateVerificationCode()
	if err != nil {
		return "", apperrors.New(op, err)
	}

	// 7. В транзакции сохраняем verification code и записываем auth_event.
	now := time.Now().UTC()
	err = s.transactions(ctx, func(repos Repositories) error {
		if err := repos.VerificationCodes.Create(ctx, &records.VerificationCode{
			ID:            uuid.New(),
			IdentityID:    &identityID,
			Email:         &identity.Email,
			Purpose:       verificationPurposeAccountDelete,
			CodeHash:      utils.HashSHA256(code),
			AttemptsCount: 0,
			MaxAttempts:   s.features.CodeMaxAttempts,
			ExpiresAt:     now.Add(s.features.CodeTTL),
			CreatedAt:     now,
		}); err != nil {
			return err
		}
		return repos.AuthEvents.Create(ctx, &records.AuthEvent{
			ID:         uuid.New(),
			IdentityID: &identityID,
			EventType:  authEventAccountDeleteStarted,
			CreatedAt:  now,
		})
	})
	if err != nil {
		return "", apperrors.New(op, err)
	}

	// 8. После commit публикуем Kafka-событие notification.email.account_delete_code.send.
	identityIDString := identityID.String()
	if err := s.events.PublishAccountDeleteCodeSend(ctx, events.VerificationCodeSendPayload{
		IdentityID: &identityIDString,
		Email:      identity.Email,
		Code:       code,
		Purpose:    verificationPurposeAccountDelete,
	}); err != nil {
		return "", apperrors.New(op, err)
	}

	// 9. Возвращаем сообщение об отправке.
	return accountDeleteCodeSentMessage, nil
}
