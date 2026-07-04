package authservice

import (
	"context"
	"crypto/rand"
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
	deviceCodeLen         = 40
	userCodeLen           = 8
	deviceCodeTTL         = 10 * time.Minute
	deviceCodeInterval    = 3
	deviceStatusPending   = "pending"
	deviceStatusConfirmed = "confirmed"
)

// DeviceStart - запуск авторизации CLI через device code grant
/*
	1. Сгенерировать device_code (длинный случайный токен) и user_code (короткий человекочитаемый).
	2. Хешировать device_code для хранения в БД.
	3. Сохранить запись в device_authorization_codes со статусом pending.
	4. Вернуть device_code, user_code, verification_url, expires_in, interval.
*/
func (s *AuthService) DeviceStart(ctx context.Context) (*DeviceStartOutput, error) {
	const op = "AuthService.DeviceStart"

	// 1. Сгенерировать device_code и user_code.
	deviceCode := utils.GenerateRandomString(deviceCodeLen)
	deviceCodeHash := utils.HashSHA256(deviceCode)
	userCode := generateUserCode()

	// 2. Сохранить запись в device_authorization_codes.
	now := time.Now().UTC()
	dac := &records.DeviceAuthorizationCode{
		ID:             uuid.New(),
		DeviceCodeHash: deviceCodeHash,
		UserCode:       userCode,
		Status:         deviceStatusPending,
		ExpiresAt:      now.Add(deviceCodeTTL),
		Interval:       deviceCodeInterval,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.deviceAuthorizationCodes.Create(ctx, dac); err != nil {
		return nil, apperrors.New(op, err)
	}

	// 3. Вернуть ответ.
	return &DeviceStartOutput{
		DeviceCode:      deviceCode,
		UserCode:        userCode,
		VerificationURL: s.verificationURL,
		ExpiresIn:       int32(deviceCodeTTL.Seconds()),
		Interval:        deviceCodeInterval,
	}, nil
}

// DeviceConfirm - подтверждение CLI-входа из браузера
/*
	1. Получить identity_id из auth-контекста (BearerAuth).
	2. Найти запись по user_code.
	3. Проверить, что статус pending и не истёк.
	4. Подтвердить: установить identity_id, статус confirmed, confirmed_at.
	5. Вернуть status: "confirmed".
*/
func (s *AuthService) DeviceConfirm(ctx context.Context, input *DeviceConfirmInput) (*DeviceConfirmOutput, error) {
	const op = "AuthService.DeviceConfirm"

	// 1. Получить identity_id из auth-контекста.
	identityID, _, err := authctx.FromContext(ctx)
	if err != nil {
		return nil, apperrors.New(op, err)
	}

	// 2. Найти запись по user_code.
	dac, err := s.deviceAuthorizationCodes.GetByUserCode(ctx, input.UserCode)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.New(op, ErrDeviceCodeNotFound)
		}
		return nil, apperrors.New(op, err)
	}

	// 3. Проверить статус и срок жизни.
	if dac.Status != deviceStatusPending {
		return nil, apperrors.New(op, ErrDeviceCodeAlreadyConfirmed)
	}
	if time.Now().After(dac.ExpiresAt) {
		return nil, apperrors.New(op, ErrDeviceCodeExpired)
	}

	// 4. Подтвердить.
	if err := s.deviceAuthorizationCodes.Confirm(ctx, dac.ID, identityID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.New(op, ErrDeviceCodeNotFound)
		}
		return nil, apperrors.New(op, err)
	}

	return &DeviceConfirmOutput{
		Status: deviceStatusConfirmed,
	}, nil
}

// DeviceToken - CLI опрашивает endpoint для получения токенов
/*
	1. Хешировать device_code и найти запись.
	2. Проверить, что запись не истекла.
	3. Если статус pending — вернуть 428 (через ErrDeviceCodeNotConfirmed).
	4. Проверить частоту опроса (last_polled_at + interval > now → ErrPollTooFrequent).
	5. Сгенерировать access_token и refresh_token.
	6. В транзакции создать session и обновить last_polled_at.
	7. Вернуть токены.
*/
func (s *AuthService) DeviceToken(ctx context.Context, input *DeviceTokenInput) (*DeviceTokenOutput, error) {
	const op = "AuthService.DeviceToken"

	if s.transactions == nil {
		return nil, apperrors.New(op, ErrTransactionsNotConfigured)
	}

	// 1. Хешировать device_code и найти запись.
	deviceCodeHash := utils.HashSHA256(input.DeviceCode)
	dac, err := s.deviceAuthorizationCodes.GetByDeviceCodeHash(ctx, deviceCodeHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.New(op, ErrDeviceCodeNotFound)
		}
		return nil, apperrors.New(op, err)
	}

	// 2. Проверить, что запись не истекла.
	if time.Now().After(dac.ExpiresAt) {
		return nil, apperrors.New(op, ErrDeviceCodeExpired)
	}

	// 3. Если статус pending — вернуть ошибку.
	if dac.Status == deviceStatusPending {
		return nil, apperrors.New(op, ErrDeviceCodeNotConfirmed)
	}

	// 4. Проверить частоту опроса.
	if dac.LastPolledAt != nil {
		elapsed := time.Since(*dac.LastPolledAt)
		if elapsed < time.Duration(dac.Interval)*time.Second {
			return nil, apperrors.New(op, ErrPollTooFrequent)
		}
	}

	if dac.IdentityID == nil {
		return nil, apperrors.New(op, ErrDeviceCodeNotFound)
	}

	// 5. Сгенерировать токены.
	sessionID := uuid.New()
	refreshToken, refreshTokenHash, err := s.tokens.GenerateRefreshToken()
	if err != nil {
		return nil, apperrors.New(op, err)
	}
	accessToken, accessExpiresAt, err := s.tokens.GenerateAccessToken(*dac.IdentityID, sessionID)
	if err != nil {
		return nil, apperrors.New(op, err)
	}

	now := time.Now().UTC()

	// 6. В транзакции создать session и обновить last_polled_at.
	err = s.transactions(ctx, func(repos Repositories) error {
		if err := repos.Sessions.Create(ctx, &records.Session{
			ID:               sessionID,
			IdentityID:       *dac.IdentityID,
			RefreshTokenHash: refreshTokenHash,
			UserAgent:        "cli",
			ExpiresAt:        now.Add(s.features.RefreshSessionTTL),
			CreatedAt:        now,
			UpdatedAt:        now,
		}); err != nil {
			return err
		}

		if err := repos.DeviceAuthorizationCodes.UpdateLastPolledAt(ctx, dac.ID); err != nil {
			return err
		}

		return repos.AuthEvents.Create(ctx, &records.AuthEvent{
			ID:         uuid.New(),
			IdentityID: dac.IdentityID,
			EventType:  "device_login",
			UserAgent:  "cli",
			CreatedAt:  now,
		})
	})
	if err != nil {
		return nil, apperrors.New(op, err)
	}

	// 7. Опубликовать событие.
	if err := s.events.PublishIdentityLogin(ctx, events.IdentityLoginPayload{
		IdentityID: dac.IdentityID.String(),
		Email:      "",
	}); err != nil {
		return nil, apperrors.New(op, err)
	}

	return &DeviceTokenOutput{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresInSeconds(accessExpiresAt),
	}, nil
}

// CliRefresh - обновление CLI access token через refresh token в body
/*
	1. Хешировать refresh token и найти session.
	2. Проверить, что сессия не отозвана и не истекла.
	3. Найти identity и проверить, что аккаунт активен.
	4. Сгенерировать новый refresh token и access token.
	5. В транзакции отозвать старую сессию, создать новую, записать auth_event.
	6. Вернуть access_token и expires_in (без refresh_token — CLI хранит его).
*/
func (s *AuthService) CliRefresh(ctx context.Context, input *CliRefreshInput) (*CliRefreshOutput, error) {
	const op = "AuthService.CliRefresh"

	if s.transactions == nil {
		return nil, apperrors.New(op, ErrTransactionsNotConfigured)
	}

	// 1. Хешировать refresh token и найти session.
	refreshTokenHash := utils.HashSHA256(input.RefreshToken)
	session, err := s.sessions.GetByRefreshTokenHash(ctx, refreshTokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.New(op, ErrInvalidRefreshToken)
		}
		return nil, apperrors.New(op, err)
	}

	// 2. Проверить, что сессия не отозвана и не истекла.
	now := time.Now().UTC()
	if session.RevokedAt != nil {
		return nil, apperrors.New(op, ErrSessionRevoked)
	}
	if !session.ExpiresAt.After(now) {
		return nil, apperrors.New(op, ErrSessionExpired)
	}

	// 3. Найти identity и проверить, что аккаунт активен.
	identity, err := s.identities.GetByID(ctx, session.IdentityID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.New(op, ErrInvalidRefreshToken)
		}
		return nil, apperrors.New(op, err)
	}
	if identity.Status != identityStatusActive {
		return nil, apperrors.New(op, ErrIdentityNotActive)
	}

	// 4. Сгенерировать новый refresh token и access token.
	newSessionID := uuid.New()
	newRefreshToken, newRefreshTokenHash, err := s.tokens.GenerateRefreshToken()
	if err != nil {
		return nil, apperrors.New(op, err)
	}
	accessToken, accessExpiresAt, err := s.tokens.GenerateAccessToken(identity.ID, newSessionID)
	if err != nil {
		return nil, apperrors.New(op, err)
	}

	// 5. В транзакции отозвать старую сессию, создать новую, записать auth_event.
	err = s.transactions(ctx, func(repos Repositories) error {
		if err := repos.Sessions.Revoke(ctx, session.ID); err != nil {
			return err
		}
		if err := repos.Sessions.Create(ctx, &records.Session{
			ID:               newSessionID,
			IdentityID:       identity.ID,
			RefreshTokenHash: newRefreshTokenHash,
			UserAgent:        session.UserAgent,
			IPAddress:        session.IPAddress,
			ExpiresAt:        now.Add(s.features.RefreshSessionTTL),
			CreatedAt:        now,
			UpdatedAt:        now,
		}); err != nil {
			return err
		}

		return repos.AuthEvents.Create(ctx, &records.AuthEvent{
			ID:         uuid.New(),
			IdentityID: &identity.ID,
			EventType:  authEventRefresh,
			IPAddress:  session.IPAddress,
			UserAgent:  session.UserAgent,
			CreatedAt:  now,
		})
	})
	if err != nil {
		return nil, apperrors.New(op, err)
	}

	return &CliRefreshOutput{
		AccessToken:  accessToken,
		ExpiresIn:    expiresInSeconds(accessExpiresAt),
		RefreshToken: newRefreshToken,
	}, nil
}

// generateUserCode генерирует короткий человекочитаемый код вида ABCD-EFGH
func generateUserCode() string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // без I, O, 0, 1 для читаемости
	code := make([]byte, 8)
	for i := range code {
		b := make([]byte, 1)
		_, _ = rand.Read(b)
		code[i] = charset[int(b[0])%len(charset)]
	}
	return string(code[:4]) + "-" + string(code[4:])
}
