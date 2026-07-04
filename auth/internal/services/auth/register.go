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
	authEventRegister           = "register"
	authEventRegisterVerified   = "register_verified"
	authEventRegisterCodeResent = "register_code_resent"
)

// Register - сервис регистрации пользователя
/*
	1. Проверить, что транзакции настроены.
	2. Сгенерировать identity_id, verification code и захешировать password.
	3. В транзакции проверить, что identity с таким email еще не существует.
	4. В транзакции создать identity со статусом pending_verification и email_verified=false.
	5. В транзакции создать credentials с password_hash.
	6. В транзакции сохранить verification code purpose=register, code_hash и expires_at.
	7. В транзакции записать auth_event регистрации.
	8. После commit опубликовать Kafka-событие identity.created (payload: identity_id, email).
	9. После commit опубликовать Kafka-событие notification.email.verification_code.send (payload: identity_id, email, code, purpose).
	10. Вернуть identity_id, email и статус pending_verification.
*/
func (s *AuthService) Register(ctx context.Context, input *RegisterInput) (*RegisterOutput, error) {
	const op = "AuthService.Register"

	// 1. Проверяем, что сервис может выполнить атомарную регистрацию.
	if s.transactions == nil {
		return nil, apperrors.New(op, ErrTransactionsNotConfigured)
	}

	// 2. Генерируем identity_id, verification code и хешируем password.
	identityID := uuid.New()
	now := time.Now().UTC()
	code, err := utils.GenerateVerificationCode()
	if err != nil {
		return nil, apperrors.New(op, err)
	}

	passwordHash, err := utils.HashPassword(input.Password)
	if err != nil {
		return nil, apperrors.New(op, err)
	}

	// В транзакции выполняем шаги 3-7.
	err = s.transactions(ctx, func(repos Repositories) error {
		// 3. Проверяем, что identity с таким email еще не существует.
		if _, err := repos.Identities.GetByEmail(ctx, input.Email); err == nil {
			return apperrors.New(op, ErrEmailAlreadyExists)
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return apperrors.New(op, err)
		}

		// 4. Создаем identity со статусом pending_verification и email_verified=false.
		if err := repos.Identities.Create(ctx, &records.Identity{
			ID:            identityID,
			Email:         input.Email,
			EmailVerified: false,
			Status:        identityStatusPendingVerification,
			CreatedAt:     now,
			UpdatedAt:     now,
		}); err != nil {
			return err
		}

		// 5. Создаем credentials с password_hash.
		if err := repos.Credentials.Create(ctx, &records.Credential{
			IdentityID:        identityID,
			PasswordHash:      passwordHash,
			PasswordChangedAt: now,
			CreatedAt:         now,
			UpdatedAt:         now,
		}); err != nil {
			return err
		}

		// 6. Сохраняем verification code purpose=register, code_hash и expires_at.
		if err := repos.VerificationCodes.Create(ctx, &records.VerificationCode{
			ID:            uuid.New(),
			IdentityID:    &identityID,
			Email:         &input.Email,
			Purpose:       verificationPurposeRegister,
			CodeHash:      utils.HashSHA256(code),
			AttemptsCount: 0,
			MaxAttempts:   s.features.CodeMaxAttempts,
			ExpiresAt:     now.Add(s.features.CodeTTL),
			CreatedAt:     now,
		}); err != nil {
			return err
		}

		// 7. Записываем auth_event регистрации.
		return repos.AuthEvents.Create(ctx, &records.AuthEvent{
			ID:         uuid.New(),
			IdentityID: &identityID,
			EventType:  authEventRegister,
			CreatedAt:  now,
		})
	})
	if err != nil {
		return nil, apperrors.New(op, err)
	}

	// 8. После commit публикуем Kafka-событие identity.created.
	identityIDString := identityID.String()
	if err := s.events.PublishIdentityCreated(ctx, events.IdentityCreatedPayload{
		IdentityID: identityIDString,
		Email:      input.Email,
	}); err != nil {
		return nil, apperrors.New(op, err)
	}

	// 9. После commit публикуем Kafka-событие для отправки verification code.
	if err := s.events.PublishVerificationCodeSend(ctx, events.VerificationCodeSendPayload{
		IdentityID: &identityIDString,
		Email:      input.Email,
		Code:       code,
		Purpose:    verificationPurposeRegister,
	}); err != nil {
		return nil, apperrors.New(op, err)
	}

	// 10. Возвращаем identity_id, email и статус pending_verification.
	return &RegisterOutput{
		IdentityID: identityIDString,
		Email:      input.Email,
		Status:     identityStatusPendingVerification,
	}, nil
}

// RegisterVerify - сервис подтверждения регистрации по коду
/*
	1. Найти активный verification code purpose=register по email.
	2. Проверить, что код не истек, не consumed и attempts_count меньше max_attempts.
	3. Сравнить переданный code с сохраненным code_hash.
	4. Если код неверный, увеличить attempts_count и вернуть ошибку.
	5. Если код верный, в транзакции отметить email_verified=true и статус identity=active.
	6. В той же транзакции пометить verification code consumed.
	7. В той же транзакции записать auth_event подтверждения регистрации.
	8. После commit опубликовать Kafka-событие identity.updated (payload: identity_id, email_verified=true, status=active).
	9. После commit опубликовать Kafka-событие profile.create (payload: identity_id, email).
	10. Вернуть сообщение об успешной верификации.
*/
func (s *AuthService) RegisterVerify(ctx context.Context, input *VerifyCodeInput) (string, error) {
	const op = "AuthService.RegisterVerify"

	if s.transactions == nil {
		return "", apperrors.New(op, ErrTransactionsNotConfigured)
	}

	// 1. Найти активный verification code purpose=register по email.
	code, err := s.verificationCodes.GetActiveByEmailAndPurpose(ctx, input.Email, verificationPurposeRegister)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", apperrors.New(op, ErrInvalidCode)
		}
		return "", apperrors.New(op, err)
	}

	// 2. Проверить, что attempts_count меньше max_attempts.
	if code.AttemptsCount >= code.MaxAttempts {
		return "", apperrors.New(op, ErrTooManyAttempts)
	}

	// 3. Сравнить переданный code с сохраненным code_hash.
	if utils.HashSHA256(input.Code) != code.CodeHash {
		// 4. Если код неверный, увеличить attempts_count и вернуть ошибку.
		if err := s.verificationCodes.IncrementAttempts(ctx, code.ID); err != nil {
			return "", apperrors.New(op, err)
		}
		return "", apperrors.New(op, ErrInvalidCode)
	}

	if code.IdentityID == nil {
		return "", apperrors.New(op, ErrInvalidCode)
	}

	identityID := *code.IdentityID
	identity, err := s.identities.GetByID(ctx, identityID)
	if err != nil {
		return "", apperrors.New(op, err)
	}
	if identity.Status != identityStatusPendingVerification || identity.EmailVerified {
		return "", apperrors.New(op, ErrInvalidCode)
	}

	err = s.transactions(ctx, func(repos Repositories) error {
		// 5. Если код верный, в транзакции отметить email_verified=true и статус identity=active.
		if err := repos.Identities.SetEmailVerified(ctx, identityID); err != nil {
			return err
		}
		if err := repos.Identities.SetStatus(ctx, identityID, identityStatusActive); err != nil {
			return err
		}

		// 6. В той же транзакции пометить verification code consumed.
		if err := repos.VerificationCodes.Consume(ctx, code.ID); err != nil {
			return err
		}

		// 7. В той же транзакции записать auth_event подтверждения регистрации.
		return repos.AuthEvents.Create(ctx, &records.AuthEvent{
			ID:         uuid.New(),
			IdentityID: &identityID,
			EventType:  authEventRegisterVerified,
			CreatedAt:  time.Now().UTC(),
		})
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", apperrors.New(op, ErrInvalidCode)
		}
		return "", apperrors.New(op, err)
	}

	identityIDString := identityID.String()
	emailVerified := true

	// 8. После commit опубликовать Kafka-событие identity.updated.
	if err := s.events.PublishIdentityUpdated(ctx, events.IdentityUpdatedPayload{
		IdentityID:    identityIDString,
		Email:         input.Email,
		Status:        identityStatusActive,
		EmailVerified: &emailVerified,
	}); err != nil {
		return "", apperrors.New(op, err)
	}

	// 10. Вернуть сообщение об успешной верификации.
	return "registration verified", nil
}

// ResendVerificationCode - сервис повторной отправки кода регистрации
/*
	1. Найти identity по email.
	2. Проверить, что аккаунт существует и еще не подтвержден.
	3. Проверить rate limit повторной отправки кода.
	4. Сгенерировать новый verification code purpose=register.
	5. В транзакции сохранить verification code и записать auth_event повторной отправки.
	6. После commit опубликовать Kafka-событие notification.email.verification_code.send (payload: identity_id, email, code, purpose).
	7. Отправка кода выполняется notification-service по Kafka-событию.
	8. Вернуть сообщение об отправке.
*/
func (s *AuthService) ResendVerificationCode(ctx context.Context, input *ResendCodeInput) (string, error) {
	const op = "AuthService.ResendVerificationCode"

	if s.transactions == nil {
		return "", apperrors.New(op, ErrTransactionsNotConfigured)
	}
	if s.rateLimiter == nil {
		return "", apperrors.New(op, ErrRateLimiterNotConfigured)
	}

	// 1. Найти identity по email.
	identity, err := s.identities.GetByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", apperrors.New(op, ErrIdentityNotFound)
		}
		return "", apperrors.New(op, err)
	}

	// 2. Проверить, что аккаунт существует и еще не подтвержден.
	if identity.EmailVerified || identity.Status != identityStatusPendingVerification {
		return "", apperrors.New(op, ErrEmailAlreadyVerified)
	}

	// 3. Проверить rate limit повторной отправки кода.
	if err := s.rateLimiter.Allow(
		ctx,
		"rate:verification_code:register:"+input.Email,
		s.features.CodeResendCooldown,
		s.features.CodeResendWindow,
		int64(s.features.CodeResendMaxRequests),
	); err != nil {
		return "", apperrors.New(op, err)
	}

	// 4. Сгенерировать новый verification code purpose=register.
	code, err := utils.GenerateVerificationCode()
	if err != nil {
		return "", apperrors.New(op, err)
	}

	now := time.Now().UTC()
	err = s.transactions(ctx, func(repos Repositories) error {
		// 5. В транзакции сохранить verification code и записать auth_event повторной отправки.
		if err := repos.VerificationCodes.Create(ctx, &records.VerificationCode{
			ID:            uuid.New(),
			IdentityID:    &identity.ID,
			Email:         &identity.Email,
			Purpose:       verificationPurposeRegister,
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
			EventType:  authEventRegisterCodeResent,
			CreatedAt:  now,
		})
	})
	if err != nil {
		return "", apperrors.New(op, err)
	}

	// 6. После commit опубликовать Kafka-событие notification.email.verification_code.send.
	identityIDString := identity.ID.String()
	if err := s.events.PublishVerificationCodeSend(ctx, events.VerificationCodeSendPayload{
		IdentityID: &identityIDString,
		Email:      identity.Email,
		Code:       code,
		Purpose:    verificationPurposeRegister,
	}); err != nil {
		return "", apperrors.New(op, err)
	}

	// 7. Отправка кода выполняется notification-service по Kafka-событию.
	// 8. Вернуть сообщение об отправке.
	return "verification code sent", nil
}
