package authservice

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"auth/internal/events"
	"auth/internal/models/records"
	apperrors "auth/pkg/errors"
	"auth/pkg/utils"
)

const (
	authEventPasswordForgot           = "password_forgot"
	authEventPasswordForgotVerified   = "password_forgot_verified"
	authEventPasswordForgotCodeResent = "password_forgot_code_resent"
	authEventPasswordReset            = "password_reset"
)

// ForgotPassword - сервис запуска восстановления пароля
/*
	1. Попробовать найти identity по email.
	2. Если identity не найдена, вернуть нейтральное сообщение без раскрытия существования email.
	3. Если аккаунт найден и active/email_verified, проверить rate limit отправки кода восстановления.
	4. Сгенерировать verification code purpose=password_forgot.
	5. В транзакции сохранить verification code и auth_event password_forgot.
	6. После commit опубликовать Kafka-событие notification.email.password_reset_code.send (payload: identity_id, email, code, purpose).
	7. Если аккаунт недоступен (not active или email не verified), вернуть нейтральное сообщение без отправки кода.
	8. Вернуть нейтральное сообщение об отправке.
*/
func (s *AuthService) ForgotPassword(ctx context.Context, input *ForgotPasswordInput) (string, error) {
	const op = "AuthService.ForgotPassword"

	// 1. Проверяем, что сервис может работать с транзакциями и rate limiter'ом.
	if s.transactions == nil {
		return "", apperrors.New(op, ErrTransactionsNotConfigured)
	}
	if s.rateLimiter == nil {
		return "", apperrors.New(op, ErrRateLimiterNotConfigured)
	}

	// 2. Пытаемся найти identity по email. Если не найдена — возвращаем нейтральное сообщение.
	identity, err := s.identities.GetByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return passwordCodeSentMessage, nil
		}
		return "", apperrors.New(op, err)
	}

	// 3. Если аккаунт не active или email не verified — возвращаем нейтральное сообщение без отправки кода.
	if identity.Status != identityStatusActive || !identity.EmailVerified {
		return passwordCodeSentMessage, nil
	}

	// 4. Проверяем rate limit отправки кода восстановления.
	if err := s.rateLimiter.Allow(ctx, "rate:verification_code:password_forgot:"+input.Email, s.features.CodeResendCooldown, s.features.CodeResendWindow, int64(s.features.CodeResendMaxRequests)); err != nil {
		return "", apperrors.New(op, err)
	}

	// 5. Генерируем verification code purpose=password_forgot.
	code, err := utils.GenerateVerificationCode()
	if err != nil {
		return "", apperrors.New(op, err)
	}

	// 6. В транзакции сохраняем verification code и auth_event.
	now := time.Now().UTC()
	err = s.transactions(ctx, func(repos Repositories) error {
		if err := repos.VerificationCodes.Create(ctx, &records.VerificationCode{
			ID:            uuid.New(),
			IdentityID:    &identity.ID,
			Email:         &identity.Email,
			Purpose:       verificationPurposePasswordForgot,
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
			IdentityID: &identity.ID,
			EventType:  authEventPasswordForgot,
			CreatedAt:  now,
		})
	})
	if err != nil {
		return "", apperrors.New(op, err)
	}

	// 7. После commit публикуем Kafka-событие notification.email.password_reset_code.send.
	identityIDString := identity.ID.String()
	if err := s.events.PublishPasswordResetCodeSend(ctx, events.VerificationCodeSendPayload{
		IdentityID: &identityIDString,
		Email:      identity.Email,
		Code:       code,
		Purpose:    verificationPurposePasswordForgot,
	}); err != nil {
		return "", apperrors.New(op, err)
	}

	// 8. Возвращаем нейтральное сообщение об отправке.
	return passwordCodeSentMessage, nil
}

// ForgotPasswordVerify - сервис проверки кода восстановления пароля
/*
	1. Найти активный verification code purpose=password_forgot по email.
	2. Проверить, что attempts_count меньше max_attempts.
	3. Сравнить переданный code с сохраненным code_hash.
	4. Если код неверный, увеличить attempts_count и вернуть ошибку.
	5. Если код верный, сгенерировать reset token и token_hash.
	6. В транзакции создать password_reset_token, пометить verification code consumed и записать auth_event password_forgot_verified.
	7. Вернуть reset_token и expires_in.
*/
func (s *AuthService) ForgotPasswordVerify(ctx context.Context, input *ForgotPasswordVerifyInput) (*ResetTokenOutput, error) {
	const op = "AuthService.ForgotPasswordVerify"

	// 1. Проверяем, что сервис может работать с транзакциями.
	if s.transactions == nil {
		return nil, apperrors.New(op, ErrTransactionsNotConfigured)
	}

	// 2. Найти активный verification code purpose=password_forgot по email.
	code, err := s.verificationCodes.GetActiveByEmailAndPurpose(ctx, input.Email, verificationPurposePasswordForgot)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.New(op, ErrInvalidCode)
		}
		return nil, apperrors.New(op, err)
	}

	// 3. Проверить, что attempts_count меньше max_attempts.
	if code.AttemptsCount >= code.MaxAttempts {
		return nil, apperrors.New(op, ErrTooManyAttempts)
	}

	// 4. Сравнить переданный code с сохраненным code_hash.
	if utils.HashSHA256(input.Code) != code.CodeHash {
		// 5. Если код неверный, увеличить attempts_count и вернуть ошибку.
		if err := s.verificationCodes.IncrementAttempts(ctx, code.ID); err != nil {
			return nil, apperrors.New(op, err)
		}
		return nil, apperrors.New(op, ErrInvalidCode)
	}

	if code.IdentityID == nil {
		return nil, apperrors.New(op, ErrInvalidCode)
	}
	identity, err := s.identities.GetByID(ctx, *code.IdentityID)
	if err != nil {
		return nil, apperrors.New(op, err)
	}
	if identity.Status != identityStatusActive || !identity.EmailVerified {
		return nil, apperrors.New(op, ErrInvalidCode)
	}

	// 6. Если код верный, генерируем reset token и token_hash.
	resetToken, resetTokenHash, err := s.tokens.GenerateRefreshToken()
	if err != nil {
		return nil, apperrors.New(op, err)
	}
	identityID := *code.IdentityID
	now := time.Now().UTC()
	expiresAt := now.Add(s.features.PasswordResetTokenTTL)

	// 7. В транзакции создаем password_reset_token, помечаем verification code consumed и записываем auth_event.
	err = s.transactions(ctx, func(repos Repositories) error {
		if err := repos.PasswordResetTokens.Create(ctx, &records.PasswordResetToken{
			ID:         uuid.New(),
			IdentityID: identityID,
			TokenHash:  resetTokenHash,
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
			EventType:  authEventPasswordForgotVerified,
			CreatedAt:  now,
		})
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.New(op, ErrInvalidCode)
		}
		return nil, apperrors.New(op, err)
	}

	// 8. Возвращаем reset_token и expires_in.
	return &ResetTokenOutput{ResetToken: resetToken, ExpiresIn: expiresInSeconds(expiresAt)}, nil
}

// ForgotPasswordCodeResend - сервис повторной отправки кода восстановления пароля
/*
	1. Попробовать найти identity по email.
	2. Если identity не найдена, вернуть нейтральное сообщение без раскрытия существования email.
	3. Если аккаунт найден и active/email_verified, проверить rate limit повторной отправки.
	4. Сгенерировать новый verification code purpose=password_forgot.
	5. В транзакции сохранить verification code и auth_event password_forgot_code_resent.
	6. После commit опубликовать Kafka-событие notification.email.password_reset_code.send (payload: identity_id, email, code, purpose).
	7. Если аккаунт недоступен, вернуть нейтральное сообщение без отправки кода.
	8. Вернуть нейтральное сообщение об отправке.
*/
func (s *AuthService) ForgotPasswordCodeResend(ctx context.Context, input *ResendCodeInput) (string, error) {
	const op = "AuthService.ForgotPasswordCodeResend"

	// 1. Проверяем, что сервис может работать с транзакциями и rate limiter'ом.
	if s.transactions == nil {
		return "", apperrors.New(op, ErrTransactionsNotConfigured)
	}
	if s.rateLimiter == nil {
		return "", apperrors.New(op, ErrRateLimiterNotConfigured)
	}

	// 2. Пытаемся найти identity по email. Если не найдена — возвращаем нейтральное сообщение.
	identity, err := s.identities.GetByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return passwordCodeSentMessage, nil
		}
		return "", apperrors.New(op, err)
	}

	// 3. Если аккаунт не active или email не verified — возвращаем нейтральное сообщение без отправки кода.
	if identity.Status != identityStatusActive || !identity.EmailVerified {
		return passwordCodeSentMessage, nil
	}

	// 4. Проверяем rate limit повторной отправки.
	if err := s.rateLimiter.Allow(ctx, "rate:verification_code:password_forgot:"+input.Email, s.features.CodeResendCooldown, s.features.CodeResendWindow, int64(s.features.CodeResendMaxRequests)); err != nil {
		return "", apperrors.New(op, err)
	}

	// 5. Генерируем новый verification code purpose=password_forgot.
	code, err := utils.GenerateVerificationCode()
	if err != nil {
		return "", apperrors.New(op, err)
	}

	// 6. В транзакции сохраняем verification code и записываем auth_event.
	now := time.Now().UTC()
	err = s.transactions(ctx, func(repos Repositories) error {
		if err := repos.VerificationCodes.Create(ctx, &records.VerificationCode{
			ID:            uuid.New(),
			IdentityID:    &identity.ID,
			Email:         &identity.Email,
			Purpose:       verificationPurposePasswordForgot,
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
			IdentityID: &identity.ID,
			EventType:  authEventPasswordForgotCodeResent,
			CreatedAt:  now,
		})
	})
	if err != nil {
		return "", apperrors.New(op, err)
	}

	// 7. После commit публикуем Kafka-событие notification.email.password_reset_code.send.
	identityIDString := identity.ID.String()
	if err := s.events.PublishPasswordResetCodeSend(ctx, events.VerificationCodeSendPayload{
		IdentityID: &identityIDString,
		Email:      identity.Email,
		Code:       code,
		Purpose:    verificationPurposePasswordForgot,
	}); err != nil {
		return "", apperrors.New(op, err)
	}

	// 8. Возвращаем нейтральное сообщение об отправке.
	return passwordCodeSentMessage, nil
}

// ResetPassword - сервис установки нового пароля по reset token
/*
	1. Посчитать hash reset_token и найти активный password_reset_token.
	2. Проверить, что token не consumed и expires_at еще не истек.
	3. Захешировать new_password.
	4. В транзакции обновить password_hash у credentials, пометить password_reset_token consumed, отозвать все refresh-сессии пользователя и записать auth_event password_reset.
	5. После commit опубликовать Kafka-событие identity.updated (payload: identity_id).
	6. Вернуть сообщение об успешном сбросе пароля.
*/
func (s *AuthService) ResetPassword(ctx context.Context, input *ResetPasswordInput) (string, error) {
	const op = "AuthService.ResetPassword"

	// 1. Проверяем, что сервис может работать с транзакциями.
	if s.transactions == nil {
		return "", apperrors.New(op, ErrTransactionsNotConfigured)
	}

	// 2. Считаем hash reset_token и ищем активный password_reset_token.
	token, err := s.passwordResetTokens.GetByTokenHash(ctx, utils.HashSHA256(input.ResetToken))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", apperrors.New(op, ErrInvalidResetToken)
		}
		return "", apperrors.New(op, err)
	}

	// 3. Проверяем, что token не consumed и expires_at еще не истек.
	if token.ConsumedAt != nil {
		return "", apperrors.New(op, ErrInvalidResetToken)
	}
	if !token.ExpiresAt.After(time.Now().UTC()) {
		return "", apperrors.New(op, ErrResetTokenExpired)
	}
	identity, err := s.identities.GetByID(ctx, token.IdentityID)
	if err != nil {
		return "", apperrors.New(op, err)
	}
	if identity.Status != identityStatusActive || !identity.EmailVerified {
		return "", apperrors.New(op, ErrInvalidResetToken)
	}

	// 4. Хешируем new_password.
	passwordHash, err := utils.HashPassword(input.NewPassword)
	if err != nil {
		return "", apperrors.New(op, err)
	}

	// 5. В транзакции обновляем password_hash, помечаем token consumed, отозвать refresh-сессии и записываем auth_event.
	now := time.Now().UTC()
	err = s.transactions(ctx, func(repos Repositories) error {
		if err := repos.Credentials.UpdatePassword(ctx, token.IdentityID, passwordHash); err != nil {
			return err
		}
		if err := repos.PasswordResetTokens.Consume(ctx, token.ID); err != nil {
			return err
		}
		if err := repos.Sessions.RevokeAllByIdentityID(ctx, token.IdentityID); err != nil {
			return err
		}
		return repos.AuthEvents.Create(ctx, &records.AuthEvent{
			ID:         uuid.New(),
			IdentityID: &token.IdentityID,
			EventType:  authEventPasswordReset,
			CreatedAt:  now,
		})
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", apperrors.New(op, ErrInvalidResetToken)
		}
		return "", apperrors.New(op, err)
	}

	// 6. После commit публикуем Kafka-событие identity.updated.
	if err := s.events.PublishIdentityUpdated(ctx, events.IdentityUpdatedPayload{IdentityID: token.IdentityID.String()}); err != nil {
		return "", apperrors.New(op, err)
	}

	// 7. Возвращаем сообщение об успешном сбросе пароля.
	return passwordResetMessage, nil
}
