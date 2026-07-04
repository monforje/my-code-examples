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
	authEventPasswordChangeStarted    = "password_change_started"
	authEventPasswordChangeVerified   = "password_change_verified"
	authEventPasswordChanged          = "password_changed"
	authEventPasswordChangeCodeResent = "password_change_code_resent"
)

// ChangePassword - сервис запуска смены пароля авторизованным пользователем
/*
	1. Получить identity_id из auth-контекста.
	2. Найти identity по identity_id и проверить, что аккаунт active и email_verified.
	3. Получить credentials по identity_id и проверить current_password через password_hash.
	4. Проверить rate limit отправки кода смены пароля.
	5. Сгенерировать verification code purpose=password_change.
	6. В транзакции сохранить verification code и auth_event password_change_started.
	7. После commit опубликовать Kafka-событие notification.email.password_change_code.send (payload: identity_id, email, code, purpose).
	8. Вернуть сообщение об отправке кода.
*/
func (s *AuthService) ChangePassword(ctx context.Context, input *ChangePasswordInput) (string, error) {
	const op = "AuthService.ChangePassword"

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

	// 4. Получаем credentials и проверяем current_password через password_hash.
	credential, err := s.credentials.GetByIdentityID(ctx, identityID)
	if err != nil {
		return "", apperrors.New(op, err)
	}
	if !utils.VerifyPassword(input.CurrentPassword, credential.PasswordHash) {
		return "", apperrors.New(op, ErrCurrentPasswordIncorrect)
	}

	// 5. Проверяем rate limit отправки кода смены пароля.
	if err := s.rateLimiter.Allow(ctx, "rate:verification_code:password_change:"+identityID.String(), s.features.CodeResendCooldown, s.features.CodeResendWindow, int64(s.features.CodeResendMaxRequests)); err != nil {
		return "", apperrors.New(op, err)
	}

	// 6. Генерируем verification code purpose=password_change.
	code, err := utils.GenerateVerificationCode()
	if err != nil {
		return "", apperrors.New(op, err)
	}

	// 7. В транзакции сохраняем verification code и auth_event.
	now := time.Now().UTC()
	err = s.transactions(ctx, func(repos Repositories) error {
		if err := repos.VerificationCodes.Create(ctx, &records.VerificationCode{
			ID:            uuid.New(),
			IdentityID:    &identityID,
			Email:         &identity.Email,
			Purpose:       verificationPurposePasswordChange,
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
			EventType:  authEventPasswordChangeStarted,
			CreatedAt:  now,
		})
	})
	if err != nil {
		return "", apperrors.New(op, err)
	}

	// 8. После commit публикуем Kafka-событие notification.email.password_change_code.send.
	identityIDString := identityID.String()
	if err := s.events.PublishPasswordChangeCodeSend(ctx, events.VerificationCodeSendPayload{
		IdentityID: &identityIDString,
		Email:      identity.Email,
		Code:       code,
		Purpose:    verificationPurposePasswordChange,
	}); err != nil {
		return "", apperrors.New(op, err)
	}

	return passwordChangeCodeSentMessage, nil
}

// ChangePasswordVerify - сервис подтверждения кода смены пароля
/*
	1. Получить identity_id из auth-контекста.
	2. Найти активный verification code purpose=password_change по identity_id.
	3. Проверить, что attempts_count меньше max_attempts.
	4. Сравнить переданный code с сохраненным code_hash.
	5. Если код неверный, увеличить attempts_count и вернуть ошибку.
	6. Если код верный, сгенерировать password_change_token и token_hash.
	7. В транзакции сохранить password_change_token, пометить verification code consumed и записать auth_event password_change_verified.
	8. Вернуть password_change_token и expires_in.
*/
func (s *AuthService) ChangePasswordVerify(ctx context.Context, input *ChangePasswordVerifyInput) (*ChangePasswordVerifyOutput, error) {
	const op = "AuthService.ChangePasswordVerify"

	// 1. Проверяем, что сервис может работать с транзакциями.
	if s.transactions == nil {
		return nil, apperrors.New(op, ErrTransactionsNotConfigured)
	}

	// 2. Получаем identity_id из auth-контекста.
	identityID, _, err := authctx.FromContext(ctx)
	if err != nil {
		return nil, apperrors.New(op, err)
	}

	// 3. Найти активный verification code purpose=password_change по identity_id.
	code, err := s.verificationCodes.GetActiveByIdentityIDAndPurpose(ctx, identityID, verificationPurposePasswordChange)
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
		// 6. Если код неверный, увеличить attempts_count и вернуть ошибку.
		if err := s.verificationCodes.IncrementAttempts(ctx, code.ID); err != nil {
			return nil, apperrors.New(op, err)
		}
		return nil, apperrors.New(op, ErrInvalidCode)
	}

	// 7. Если код верный, генерируем password_change_token и token_hash.
	changeToken, changeTokenHash, err := s.tokens.GenerateRefreshToken()
	if err != nil {
		return nil, apperrors.New(op, err)
	}
	now := time.Now().UTC()
	expiresAt := now.Add(s.features.PasswordChangeTokenTTL)

	// 8. В транзакции сохраняем password_change_token, помечаем verification code consumed и записываем auth_event.
	err = s.transactions(ctx, func(repos Repositories) error {
		if err := repos.PasswordChangeTokens.Create(ctx, &records.PasswordChangeToken{
			ID:         uuid.New(),
			IdentityID: identityID,
			TokenHash:  changeTokenHash,
			ExpiresAt:  expiresAt,
			CreatedAt:  now,
		}); err != nil {
			return err
		}
		if err := repos.VerificationCodes.Consume(ctx, code.ID); err != nil {
			return err
		}
		return repos.AuthEvents.Create(ctx, &records.AuthEvent{
			ID:         uuid.New(),
			IdentityID: &identityID,
			EventType:  authEventPasswordChangeVerified,
			CreatedAt:  now,
		})
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.New(op, ErrInvalidCode)
		}
		return nil, apperrors.New(op, err)
	}

	// 9. Возвращаем password_change_token и expires_in.
	return &ChangePasswordVerifyOutput{ChangeToken: changeToken, ExpiresIn: expiresInSeconds(expiresAt)}, nil
}

// CompletePasswordChange - сервис установки нового пароля по password_change_token
/*
	1. Получить identity_id из auth-контекста.
	2. Посчитать hash change_token и найти активный password_change_token.
	3. Проверить, что token принадлежит текущему identity.
	4. Найти identity и проверить, что аккаунт active и email_verified.
	5. Захешировать new_password.
	6. В транзакции обновить password_hash, пометить token consumed, отозвать refresh-сессии и записать auth_event password_changed.
	7. После commit опубликовать Kafka-событие identity.updated (payload: identity_id).
	8. Вернуть сообщение об успешной смене пароля.
*/
func (s *AuthService) CompletePasswordChange(ctx context.Context, input *CompletePasswordChangeInput) (string, error) {
	const op = "AuthService.CompletePasswordChange"

	// 1. Проверяем, что сервис может работать с транзакциями.
	if s.transactions == nil {
		return "", apperrors.New(op, ErrTransactionsNotConfigured)
	}

	// 2. Получаем identity_id из auth-контекста.
	identityID, _, err := authctx.FromContext(ctx)
	if err != nil {
		return "", apperrors.New(op, err)
	}

	// 3. Считаем hash change_token и ищем активный password_change_token.
	token, err := s.passwordChangeTokens.GetByTokenHash(ctx, utils.HashSHA256(input.ChangeToken))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", apperrors.New(op, ErrInvalidChangeToken)
		}
		return "", apperrors.New(op, err)
	}

	// 4. Проверяем, что token принадлежит текущему identity.
	if token.IdentityID != identityID {
		return "", apperrors.New(op, ErrInvalidChangeToken)
	}

	// 5. Находим identity и проверяем, что аккаунт active и email_verified.
	identity, err := s.identities.GetByID(ctx, identityID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", apperrors.New(op, ErrInvalidChangeToken)
		}
		return "", apperrors.New(op, err)
	}
	if identity.Status != identityStatusActive {
		return "", apperrors.New(op, ErrIdentityNotActive)
	}
	if !identity.EmailVerified {
		return "", apperrors.New(op, ErrEmailNotVerified)
	}

	// 6. Хешируем new_password.
	passwordHash, err := utils.HashPassword(input.NewPassword)
	if err != nil {
		return "", apperrors.New(op, err)
	}

	// 7. В транзакции обновляем password_hash, помечаем token consumed, отозвать refresh-сессии и записываем auth_event.
	now := time.Now().UTC()
	err = s.transactions(ctx, func(repos Repositories) error {
		if err := repos.Credentials.UpdatePassword(ctx, identityID, passwordHash); err != nil {
			return err
		}
		if err := repos.PasswordChangeTokens.Consume(ctx, token.ID); err != nil {
			return err
		}
		if err := repos.Sessions.RevokeAllByIdentityID(ctx, identityID); err != nil {
			return err
		}
		return repos.AuthEvents.Create(ctx, &records.AuthEvent{
			ID:         uuid.New(),
			IdentityID: &identityID,
			EventType:  authEventPasswordChanged,
			CreatedAt:  now,
		})
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", apperrors.New(op, ErrInvalidChangeToken)
		}
		return "", apperrors.New(op, err)
	}

	// 8. После commit публикуем Kafka-событие identity.updated.
	if err := s.events.PublishIdentityUpdated(ctx, events.IdentityUpdatedPayload{IdentityID: identityID.String()}); err != nil {
		return "", apperrors.New(op, err)
	}

	// 9. Возвращаем сообщение об успешной смене пароля.
	return passwordChangedMessage, nil
}

// ChangePasswordCodeResend - сервис повторной отправки кода смены пароля
/*
	1. Получить identity_id из auth-контекста.
	2. Найти identity по identity_id и проверить, что аккаунт active и email_verified.
	3. Проверить, что существует активный verification code purpose=password_change.
	4. Проверить rate limit повторной отправки.
	5. Сгенерировать новый verification code purpose=password_change.
	6. В транзакции сохранить verification code и записать auth_event password_change_code_resent.
	7. После commit опубликовать Kafka-событие notification.email.password_change_code.send (payload: identity_id, email, code, purpose).
	8. Вернуть сообщение об отправке.
*/
func (s *AuthService) ChangePasswordCodeResend(ctx context.Context) (string, error) {
	const op = "AuthService.ChangePasswordCodeResend"

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

	// 4. Проверяем, что существует активный verification code purpose=password_change.
	if _, err := s.verificationCodes.GetActiveByIdentityIDAndPurpose(ctx, identityID, verificationPurposePasswordChange); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", apperrors.New(op, ErrPasswordChangeNotFound)
		}
		return "", apperrors.New(op, err)
	}

	// 5. Проверяем rate limit повторной отправки.
	if err := s.rateLimiter.Allow(ctx, "rate:verification_code:password_change:"+identityID.String(), s.features.CodeResendCooldown, s.features.CodeResendWindow, int64(s.features.CodeResendMaxRequests)); err != nil {
		return "", apperrors.New(op, err)
	}

	// 6. Генерируем новый verification code purpose=password_change.
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
			Purpose:       verificationPurposePasswordChange,
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
			EventType:  authEventPasswordChangeCodeResent,
			CreatedAt:  now,
		})
	})
	if err != nil {
		return "", apperrors.New(op, err)
	}

	// 8. После commit публикуем Kafka-событие notification.email.password_change_code.send.
	identityIDString := identityID.String()
	if err := s.events.PublishPasswordChangeCodeSend(ctx, events.VerificationCodeSendPayload{
		IdentityID: &identityIDString,
		Email:      identity.Email,
		Code:       code,
		Purpose:    verificationPurposePasswordChange,
	}); err != nil {
		return "", apperrors.New(op, err)
	}

	// 9. Возвращаем сообщение об отправке.
	return passwordChangeCodeSentMessage, nil
}
