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
	authEventEmailChangeStarted    = "email_change_started"
	authEventEmailChangeVerified   = "email_change_verified"
	authEventEmailChangeConfirmed  = "email_change_confirmed"
	authEventEmailChanged          = "email_changed"
	authEventEmailChangeCodeResent = "email_change_code_resent"
)

// ChangeEmail - сервис запуска смены email (шаг 1)
/*
	1. Получить identity_id из auth-контекста.
	2. Найти identity по identity_id и проверить, что аккаунт active и email_verified.
	3. Получить credentials по identity_id и проверить password через password_hash.
	4. Проверить rate limit отправки кода смены email.
	5. Сгенерировать verification code purpose=email_change_current.
	6. В транзакции создать email_change_request (status=pending, new_email=""), сохранить verification code и записать auth_event.
	7. После commit опубликовать Kafka-событие notification.email.email_change_code.send (payload: identity_id, current_email, code, purpose=email_change_current).
	8. Вернуть сообщение об отправке кода.
*/
func (s *AuthService) ChangeEmail(ctx context.Context, input *ChangeEmailInput) (string, error) {
	const op = "AuthService.ChangeEmail"

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

	// 5. Проверяем rate limit отправки кода смены email.
	if err := s.rateLimiter.Allow(ctx, "rate:verification_code:email_change_current:"+identityID.String(), s.features.CodeResendCooldown, s.features.CodeResendWindow, int64(s.features.CodeResendMaxRequests)); err != nil {
		return "", apperrors.New(op, err)
	}

	// 6. Генерируем verification code purpose=email_change_current.
	code, err := utils.GenerateVerificationCode()
	if err != nil {
		return "", apperrors.New(op, err)
	}

	// 7. В транзакции создаем email_change_request, сохраняем verification code и записываем auth_event.
	now := time.Now().UTC()
	expiresAt := now.Add(s.features.EmailChangeTokenTTL)
	err = s.transactions(ctx, func(repos Repositories) error {
		if err := repos.EmailChangeRequests.Create(ctx, &records.EmailChangeRequest{
			ID:         uuid.New(),
			IdentityID: identityID,
			Status:     emailChangeStatusPending,
			ExpiresAt:  expiresAt,
			CreatedAt:  now,
			UpdatedAt:  now,
		}); err != nil {
			return err
		}

		if err := repos.VerificationCodes.Create(ctx, &records.VerificationCode{
			ID:            uuid.New(),
			IdentityID:    &identityID,
			Email:         &identity.Email,
			Purpose:       verificationPurposeEmailChangeCurrent,
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
			EventType:  authEventEmailChangeStarted,
			CreatedAt:  now,
		})
	})
	if err != nil {
		return "", apperrors.New(op, err)
	}

	// 8. После commit публикуем Kafka-событие notification.email.email_change_code.send.
	identityIDString := identityID.String()
	if err := s.events.PublishEmailChangeCodeSend(ctx, events.VerificationCodeSendPayload{
		IdentityID: &identityIDString,
		Email:      identity.Email,
		Code:       code,
		Purpose:    verificationPurposeEmailChangeCurrent,
	}); err != nil {
		return "", apperrors.New(op, err)
	}

	// 9. Возвращаем сообщение об отправке кода.
	return emailChangeCodeSentMessage, nil
}

// ChangeEmailVerify - сервис подтверждения текущего email кодом (шаг 2)
/*
	1. Получить identity_id из auth-контекста.
	2. Найти активный verification code purpose=email_change_current по identity_id.
	3. Проверить, что attempts_count меньше max_attempts.
	4. Сравнить переданный code с сохраненным code_hash.
	5. Если код неверный, увеличить attempts_count и вернуть ошибку.
	6. Если код верный, сгенерировать identity_token и identity_token_hash.
	7. Найти активный email_change_request (status=pending).
	8. В транзакции сохранить identity_token_hash в email_change_request, установить status=verified, пометить verification code consumed и записать auth_event.
	9. Вернуть identity_token и expires_in.
*/
func (s *AuthService) ChangeEmailVerify(ctx context.Context, input *ChangeEmailVerifyInput) (*ChangeEmailVerifyOutput, error) {
	const op = "AuthService.ChangeEmailVerify"

	// 1. Проверяем, что сервис может работать с транзакциями.
	if s.transactions == nil {
		return nil, apperrors.New(op, ErrTransactionsNotConfigured)
	}

	// 2. Получаем identity_id из auth-контекста.
	identityID, _, err := authctx.FromContext(ctx)
	if err != nil {
		return nil, apperrors.New(op, err)
	}

	// 3. Найти активный verification code purpose=email_change_current по identity_id.
	code, err := s.verificationCodes.GetActiveByIdentityIDAndPurpose(ctx, identityID, verificationPurposeEmailChangeCurrent)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.New(op, ErrInvalidCode)
		}
		return nil, apperrors.New(op, err)
	}

	// 4. Проверить, что attempts_count меньше max_attempts.
	if code.AttemptsCount >= code.MaxAttempts {
		return nil, apperrors.New(op, ErrTooManyAttempts)
	}

	// 5. Сравнить переданный code с сохраненным code_hash.
	if utils.HashSHA256(input.Code) != code.CodeHash {
		// 5.1 Если код неверный, увеличить attempts_count и вернуть ошибку.
		if err := s.verificationCodes.IncrementAttempts(ctx, code.ID); err != nil {
			return nil, apperrors.New(op, err)
		}
		return nil, apperrors.New(op, ErrInvalidCode)
	}

	// 6. Если код верный, генерируем identity_token и identity_token_hash.
	identityToken, identityTokenHash, err := s.tokens.GenerateRefreshToken()
	if err != nil {
		return nil, apperrors.New(op, err)
	}
	now := time.Now().UTC()

	// 7. Находим активный email_change_request (status=pending).
	request, err := s.emailChangeRequests.GetActiveByIdentityIDAndStatus(ctx, identityID, emailChangeStatusPending)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.New(op, ErrInvalidEmailChangeToken)
		}
		return nil, apperrors.New(op, err)
	}

	// 8. В транзакции сохраняем identity_token_hash, устанавливаем status=verified, помечаем verification code consumed и записываем auth_event.
	err = s.transactions(ctx, func(repos Repositories) error {
		if err := repos.EmailChangeRequests.UpdateTokenHash(ctx, request.ID, identityTokenHash); err != nil {
			return err
		}
		if err := repos.EmailChangeRequests.SetStatus(ctx, request.ID, emailChangeStatusVerified); err != nil {
			return err
		}
		if err := repos.VerificationCodes.Consume(ctx, code.ID); err != nil {
			return err
		}
		return repos.AuthEvents.Create(ctx, &records.AuthEvent{
			ID:         uuid.New(),
			IdentityID: &identityID,
			EventType:  authEventEmailChangeVerified,
			CreatedAt:  now,
		})
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.New(op, ErrInvalidCode)
		}
		return nil, apperrors.New(op, err)
	}

	// 9. Возвращаем identity_token и expires_in.
	expiresAt := request.ExpiresAt
	if expiresAt.IsZero() {
		expiresAt = now.Add(s.features.EmailChangeTokenTTL)
	}
	return &ChangeEmailVerifyOutput{IdentityToken: identityToken, ExpiresIn: expiresInSeconds(expiresAt)}, nil
}

// ChangeEmailConfirm - сервис подтверждения нового email (шаг 3)
/*
	1. Получить identity_id из auth-контекста.
	2. Посчитать hash identity_token и найти активный email_change_request (status=verified).
	3. Проверить, что token принадлежит текущему identity.
	4. Проверить, что new_email еще не занят другой identity.
	5. Проверить rate limit отправки кода подтверждения нового email.
	6. Сгенерировать verification code purpose=email_change_new.
	7. В транзакции обновить new_email и status=confirming в email_change_request, сохранить verification code и записать auth_event.
	8. После commit опубликовать Kafka-событие notification.email.email_change_code.send (payload: identity_id, new_email, code, purpose=email_change_new).
	9. Вернуть сообщение об отправке кода.
*/
func (s *AuthService) ChangeEmailConfirm(ctx context.Context, input *ChangeEmailConfirmInput) (string, error) {
	const op = "AuthService.ChangeEmailConfirm"

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

	// 3. Считаем hash identity_token и ищем активный email_change_request (status=verified).
	request, err := s.emailChangeRequests.GetByTokenHash(ctx, utils.HashSHA256(input.IdentityToken))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", apperrors.New(op, ErrInvalidEmailChangeToken)
		}
		return "", apperrors.New(op, err)
	}

	// 4. Проверяем, что token принадлежит текущему identity.
	if request.IdentityID != identityID {
		return "", apperrors.New(op, ErrInvalidEmailChangeToken)
	}

	// 5. Проверяем, что status=verified.
	if request.Status != emailChangeStatusVerified {
		return "", apperrors.New(op, ErrInvalidEmailChangeToken)
	}

	// 6. Проверяем, что new_email еще не занят другой identity.
	existing, err := s.identities.GetByEmail(ctx, input.NewEmail)
	if err == nil && existing != nil {
		return "", apperrors.New(op, ErrEmailAlreadyExists)
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return "", apperrors.New(op, err)
	}

	// 7. Проверяем rate limit отправки кода подтверждения нового email.
	if err := s.rateLimiter.Allow(ctx, "rate:verification_code:email_change_new:"+identityID.String(), s.features.CodeResendCooldown, s.features.CodeResendWindow, int64(s.features.CodeResendMaxRequests)); err != nil {
		return "", apperrors.New(op, err)
	}

	// 8. Генерируем verification code purpose=email_change_new.
	code, err := utils.GenerateVerificationCode()
	if err != nil {
		return "", apperrors.New(op, err)
	}

	// 9. В транзакции обновляем new_email и status, сохраняем verification code и записываем auth_event.
	now := time.Now().UTC()
	err = s.transactions(ctx, func(repos Repositories) error {
		if err := repos.EmailChangeRequests.UpdateNewEmailAndStatus(ctx, request.ID, input.NewEmail, emailChangeStatusConfirming); err != nil {
			return err
		}

		if err := repos.VerificationCodes.Create(ctx, &records.VerificationCode{
			ID:            uuid.New(),
			IdentityID:    &identityID,
			Email:         &input.NewEmail,
			Purpose:       verificationPurposeEmailChangeNew,
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
			EventType:  authEventEmailChangeConfirmed,
			CreatedAt:  now,
		})
	})
	if err != nil {
		return "", apperrors.New(op, err)
	}

	// 10. После commit публикуем Kafka-событие notification.email.email_change_code.send.
	identityIDString := identityID.String()
	if err := s.events.PublishEmailChangeCodeSend(ctx, events.VerificationCodeSendPayload{
		IdentityID: &identityIDString,
		Email:      input.NewEmail,
		Code:       code,
		Purpose:    verificationPurposeEmailChangeNew,
	}); err != nil {
		return "", apperrors.New(op, err)
	}

	// 11. Возвращаем сообщение об отправке кода.
	return emailChangeCodeSentMessage, nil
}

// ChangeEmailComplete - сервис завершения смены email по коду с нового email (шаг 4)
/*
	1. Получить identity_id из auth-контекста.
	2. Найти активный verification code purpose=email_change_new по identity_id.
	3. Проверить, что attempts_count меньше max_attempts.
	4. Сравнить переданный code с сохраненным code_hash.
	5. Если код неверный, увеличить attempts_count и вернуть ошибку.
	6. Если код верный, найти активный email_change_request (status=confirming).
	7. В транзакции обновить email у identity, пометить email_change_request consumed, отозвать все остальные сессии и записать auth_event email_changed.
	8. После commit опубликовать Kafka-событие identity.updated (payload: identity_id, email) и profile.update (payload: identity_id, email).
	9. Вернуть сообщение об успешной смене email.
*/
func (s *AuthService) ChangeEmailComplete(ctx context.Context, input *ChangeEmailCompleteInput) (string, error) {
	const op = "AuthService.ChangeEmailComplete"

	// 1. Проверяем, что сервис может работать с транзакциями.
	if s.transactions == nil {
		return "", apperrors.New(op, ErrTransactionsNotConfigured)
	}

	// 2. Получаем identity_id из auth-контекста.
	identityID, _, err := authctx.FromContext(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", apperrors.New(op, ErrInvalidCode)
		}
		return "", apperrors.New(op, err)
	}

	// 3. Найти активный verification code purpose=email_change_new по identity_id.
	code, err := s.verificationCodes.GetActiveByIdentityIDAndPurpose(ctx, identityID, verificationPurposeEmailChangeNew)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", apperrors.New(op, ErrInvalidCode)
		}
		return "", apperrors.New(op, err)
	}

	// 4. Проверить, что attempts_count меньше max_attempts.
	if code.AttemptsCount >= code.MaxAttempts {
		return "", apperrors.New(op, ErrTooManyAttempts)
	}

	// 5. Сравнить переданный code с сохраненным code_hash.
	if utils.HashSHA256(input.Code) != code.CodeHash {
		// 5.1 Если код неверный, увеличить attempts_count и вернуть ошибку.
		if err := s.verificationCodes.IncrementAttempts(ctx, code.ID); err != nil {
			return "", apperrors.New(op, err)
		}
		return "", apperrors.New(op, ErrInvalidCode)
	}

	// 6. Если код верный, находим активный email_change_request (status=confirming).
	request, err := s.emailChangeRequests.GetActiveByIdentityIDAndStatus(ctx, identityID, emailChangeStatusConfirming)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", apperrors.New(op, ErrInvalidEmailChangeToken)
		}
		return "", apperrors.New(op, err)
	}

	// 7. Находим identity и проверяем, что аккаунт active и email_verified.
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

	// 8. В транзакции обновляем email, потребляем request, отозваем остальные сессии и записываем auth_event.
	now := time.Now().UTC()
	err = s.transactions(ctx, func(repos Repositories) error {
		identity.Email = request.NewEmail
		identity.UpdatedAt = now
		if err := repos.Identities.Update(ctx, identity); err != nil {
			return err
		}
		if err := repos.EmailChangeRequests.Consume(ctx, request.ID); err != nil {
			return err
		}
		if err := repos.VerificationCodes.Consume(ctx, code.ID); err != nil {
			return err
		}
		if err := repos.Sessions.RevokeAllByIdentityID(ctx, identityID); err != nil {
			return err
		}
		return repos.AuthEvents.Create(ctx, &records.AuthEvent{
			ID:         uuid.New(),
			IdentityID: &identityID,
			EventType:  authEventEmailChanged,
			CreatedAt:  now,
		})
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", apperrors.New(op, ErrInvalidCode)
		}
		return "", apperrors.New(op, err)
	}

	// 9. После commit публикуем Kafka-событие identity.updated.
	if err := s.events.PublishIdentityUpdated(ctx, events.IdentityUpdatedPayload{
		IdentityID: identityID.String(),
		Email:      request.NewEmail,
	}); err != nil {
		return "", apperrors.New(op, err)
	}

	// 11. Возвращаем сообщение об успешной смене email.
	return emailChangedMessage, nil
}

// ChangeEmailCodeResend - сервис повторной отправки кода смены email
/*
	1. Получить identity_id из auth-контекста.
	2. Найти identity по identity_id и проверить, что аккаунт active и email_verified.
	3. Определить purpose по step: "current" → email_change_current, "new" → email_change_new.
	4. Найти активный email_change_request по identity_id и статусу.
	5. Проверить rate limit повторной отправки.
	6. Сгенерировать новый verification code.
	7. В транзакции сохранить verification code и записать auth_event.
	8. После commit опубликовать Kafka-событие notification.email.email_change_code.send.
	9. Вернуть сообщение об отправке.
*/
func (s *AuthService) ChangeEmailCodeResend(ctx context.Context, input *ChangeEmailResendInput) (string, error) {
	const op = "AuthService.ChangeEmailCodeResend"

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

	// 3. Находим identity и проверяем, что аккаунт active и email_verified.
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

	// 4. Определяем purpose и requiredStatus по step.
	var purpose string
	var requiredStatus string
	var sendEmail string
	switch input.Step {
	case "current":
		purpose = verificationPurposeEmailChangeCurrent
		requiredStatus = emailChangeStatusPending
		sendEmail = identity.Email
	case "new":
		purpose = verificationPurposeEmailChangeNew
		requiredStatus = emailChangeStatusConfirming
		// Для step=new email берём из активного request.
		requestForEmail, err := s.emailChangeRequests.GetActiveByIdentityIDAndStatus(ctx, identityID, requiredStatus)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return "", apperrors.New(op, ErrInvalidEmailChangeToken)
			}
			return "", apperrors.New(op, err)
		}
		sendEmail = requestForEmail.NewEmail
	default:
		return "", apperrors.New(op, ErrInvalidCode)
	}

	// 5. Проверяем, что существует активный email_change_request.
	if _, err := s.emailChangeRequests.GetActiveByIdentityIDAndStatus(ctx, identityID, requiredStatus); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", apperrors.New(op, ErrInvalidEmailChangeToken)
		}
		return "", apperrors.New(op, err)
	}

	// 6. Проверяем rate limit повторной отправки.
	rateLimitKey := "rate:verification_code:" + purpose + ":" + identityID.String()
	if err := s.rateLimiter.Allow(ctx, rateLimitKey, s.features.CodeResendCooldown, s.features.CodeResendWindow, int64(s.features.CodeResendMaxRequests)); err != nil {
		return "", apperrors.New(op, err)
	}

	// 7. Генерируем новый verification code.
	code, err := utils.GenerateVerificationCode()
	if err != nil {
		return "", apperrors.New(op, err)
	}

	// 8. В транзакции сохраняем verification code и записываем auth_event.
	now := time.Now().UTC()
	err = s.transactions(ctx, func(repos Repositories) error {
		if err := repos.VerificationCodes.Create(ctx, &records.VerificationCode{
			ID:            uuid.New(),
			IdentityID:    &identityID,
			Email:         &sendEmail,
			Purpose:       purpose,
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
			EventType:  authEventEmailChangeCodeResent,
			CreatedAt:  now,
		})
	})
	if err != nil {
		return "", apperrors.New(op, err)
	}

	// 9. После commit публикуем Kafka-событие notification.email.email_change_code.send.
	identityIDString := identityID.String()
	if err := s.events.PublishEmailChangeCodeSend(ctx, events.VerificationCodeSendPayload{
		IdentityID: &identityIDString,
		Email:      sendEmail,
		Code:       code,
		Purpose:    purpose,
	}); err != nil {
		return "", apperrors.New(op, err)
	}

	// 10. Возвращаем сообщение об отправке.
	return emailChangeCodeSentMessage, nil
}
