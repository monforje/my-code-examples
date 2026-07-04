package authservice

import (
	"context"
	"errors"
	"net"
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
	authEventLogin   = "login"
	authEventRefresh = "refresh"
	authEventLogout  = "logout"
)

// Login - сервис входа пользователя в аккаунт
/*
	1. Найти identity по email.
	2. Проверить, что аккаунт active и email подтвержден.
	3. Получить credentials по identity_id и сравнить password с password_hash.
	4. Сгенерировать refresh token и access token.
	5. В транзакции создать refresh-сессию и записать auth_event входа.
	6. После commit опубликовать Kafka-событие identity.login (payload: identity_id, email).
	7. Вернуть access_token, refresh_token и expires_in.
*/
func (s *AuthService) Login(ctx context.Context, input *LoginInput) (*LoginOutput, error) {
	const op = "AuthService.Login"

	if s.transactions == nil {
		return nil, apperrors.New(op, ErrTransactionsNotConfigured)
	}

	// 1. Найти identity по email.
	identity, err := s.identities.GetByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.New(op, ErrInvalidCredentials)
		}
		return nil, apperrors.New(op, err)
	}

	// 2. Проверить, что аккаунт active и email подтвержден.
	if identity.Status != identityStatusActive {
		return nil, apperrors.New(op, ErrIdentityNotActive)
	}
	if !identity.EmailVerified {
		return nil, apperrors.New(op, ErrEmailNotVerified)
	}

	// 3. Получить credentials по identity_id и сравнить password с password_hash.
	credential, err := s.credentials.GetByIdentityID(ctx, identity.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.New(op, ErrInvalidCredentials)
		}
		return nil, apperrors.New(op, err)
	}
	if !utils.VerifyPassword(input.Password, credential.PasswordHash) {
		return nil, apperrors.New(op, ErrInvalidCredentials)
	}

	// 4. Сгенерировать refresh token и access token.
	sessionID := uuid.New()
	refreshToken, refreshTokenHash, err := s.tokens.GenerateRefreshToken()
	if err != nil {
		return nil, apperrors.New(op, err)
	}
	accessToken, accessExpiresAt, err := s.tokens.GenerateAccessToken(identity.ID, sessionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.New(op, ErrInvalidRefreshToken)
		}
		return nil, apperrors.New(op, err)
	}

	now := time.Now().UTC()
	userAgent := input.UserAgent
	ipAddress := ipPointer(input.IPAddress)

	// 5. В транзакции создать refresh-сессию и записать auth_event входа.
	err = s.transactions(ctx, func(repos Repositories) error {
		if err := repos.Sessions.Create(ctx, &records.Session{
			ID:               sessionID,
			IdentityID:       identity.ID,
			RefreshTokenHash: refreshTokenHash,
			UserAgent:        userAgent,
			IPAddress:        ipAddress,
			ExpiresAt:        now.Add(s.features.RefreshSessionTTL),
			CreatedAt:        now,
			UpdatedAt:        now,
		}); err != nil {
			return err
		}

		return repos.AuthEvents.Create(ctx, &records.AuthEvent{
			ID:         uuid.New(),
			IdentityID: &identity.ID,
			EventType:  authEventLogin,
			IPAddress:  ipAddress,
			UserAgent:  userAgent,
			CreatedAt:  now,
		})
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.New(op, ErrInvalidRefreshToken)
		}
		return nil, apperrors.New(op, err)
	}

	// 6. После commit опубликовать Kafka-событие identity.login.
	if err := s.events.PublishIdentityLogin(ctx, events.IdentityLoginPayload{
		IdentityID: identity.ID.String(),
		Email:      identity.Email,
	}); err != nil {
		return nil, apperrors.New(op, err)
	}

	// 7. Вернуть access_token, refresh_token и expires_in.
	return &LoginOutput{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresInSeconds(accessExpiresAt),
	}, nil
}

// Refresh - сервис обновления токенов по refresh token
/*
	1. Получить refresh token из RefreshInput и посчитать его hash.
	2. Найти session по refresh_token_hash.
	3. Проверить, что сессия не отозвана и expires_at еще не истек.
	4. Найти identity и проверить, что аккаунт активен.
	5. Сгенерировать новый refresh token и новый access token.
	6. В транзакции отозвать текущую refresh-сессию, создать новую refresh-сессию и записать auth_event refresh.
	7. Вернуть access_token, refresh_token и expires_in.
*/
func (s *AuthService) Refresh(ctx context.Context, input *RefreshInput) (*RefreshOutput, error) {
	const op = "AuthService.Refresh"

	if s.transactions == nil {
		return nil, apperrors.New(op, ErrTransactionsNotConfigured)
	}

	// 1. Получить refresh token из RefreshInput и посчитать его hash.
	refreshTokenHash := utils.HashSHA256(input.RefreshToken)

	// 2. Найти session по refresh_token_hash.
	session, err := s.sessions.GetByRefreshTokenHash(ctx, refreshTokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.New(op, ErrInvalidRefreshToken)
		}
		return nil, apperrors.New(op, err)
	}

	// 3. Проверить, что сессия не отозвана и expires_at еще не истек.
	now := time.Now().UTC()
	if session.RevokedAt != nil {
		return nil, apperrors.New(op, ErrSessionRevoked)
	}
	if !session.ExpiresAt.After(now) {
		return nil, apperrors.New(op, ErrSessionExpired)
	}

	// 4. Найти identity и проверить, что аккаунт активен.
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

	// 5. Сгенерировать новый refresh token и новый access token.
	newSessionID := uuid.New()
	newRefreshToken, newRefreshTokenHash, err := s.tokens.GenerateRefreshToken()
	if err != nil {
		return nil, apperrors.New(op, err)
	}
	accessToken, accessExpiresAt, err := s.tokens.GenerateAccessToken(identity.ID, newSessionID)
	if err != nil {
		return nil, apperrors.New(op, err)
	}

	// 6. В транзакции отозвать текущую refresh-сессию, создать новую refresh-сессию и записать auth_event refresh.
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
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.New(op, ErrInvalidRefreshToken)
		}
		return nil, apperrors.New(op, err)
	}

	// 7. Вернуть access_token, refresh_token и expires_in.
	return &RefreshOutput{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    expiresInSeconds(accessExpiresAt),
	}, nil
}

// Logout - сервис выхода пользователя из аккаунта
/*
	1. Получить identity_id и session_id из auth-контекста.
	2. Найти активную session по id.
	3. В транзакции проставить revoked_at для session и записать auth_event выхода.
	4. После commit опубликовать Kafka-событие identity.logout (payload: identity_id).
	5. Вернуть nil, потому что HTTP-слой отдаст 204 No Content.
*/
func (s *AuthService) Logout(ctx context.Context) error {
	const op = "AuthService.Logout"

	if s.transactions == nil {
		return apperrors.New(op, ErrTransactionsNotConfigured)
	}

	// 1. Получить identity_id и session_id из auth-контекста.
	identityID, sessionID, err := authctx.FromContext(ctx)
	if err != nil {
		return apperrors.New(op, err)
	}

	// 2. Найти активную session по id.
	session, err := s.sessions.GetByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.New(op, ErrInvalidRefreshToken)
		}
		return apperrors.New(op, err)
	}
	if session.IdentityID != identityID {
		return apperrors.New(op, ErrInvalidRefreshToken)
	}
	if session.RevokedAt != nil {
		return apperrors.New(op, ErrSessionRevoked)
	}

	now := time.Now().UTC()

	// 3. В транзакции проставить revoked_at для session и записать auth_event выхода.
	err = s.transactions(ctx, func(repos Repositories) error {
		if err := repos.Sessions.Revoke(ctx, sessionID); err != nil {
			return err
		}

		return repos.AuthEvents.Create(ctx, &records.AuthEvent{
			ID:         uuid.New(),
			IdentityID: &identityID,
			EventType:  authEventLogout,
			IPAddress:  session.IPAddress,
			UserAgent:  session.UserAgent,
			CreatedAt:  now,
		})
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.New(op, ErrSessionRevoked)
		}
		return apperrors.New(op, err)
	}

	// 4. После commit опубликовать Kafka-событие identity.logout.
	if err := s.events.PublishIdentityLogout(ctx, events.IdentityLogoutPayload{
		IdentityID: identityID.String(),
	}); err != nil {
		return apperrors.New(op, err)
	}

	// 5. Вернуть nil, потому что HTTP-слой отдаст 204 No Content.
	return nil
}

func ipPointer(ip net.IP) *net.IP {
	if len(ip) == 0 {
		return nil
	}
	copyIP := append(net.IP(nil), ip...)
	return &copyIP
}

func expiresInSeconds(expiresAt time.Time) int32 {
	seconds := int32(time.Until(expiresAt).Seconds())
	if seconds < 0 {
		return 0
	}
	return seconds
}
