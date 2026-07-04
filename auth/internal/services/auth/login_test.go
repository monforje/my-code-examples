package authservice_test

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/mock/gomock"

	"auth/internal/authctx"
	"auth/internal/events"
	"auth/internal/models/records"
	authservice "auth/internal/services/auth"
	"auth/pkg/utils"
)

func TestAuthService_Login_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()
	identityID := uuid.New()
	email := "user@example.com"
	password := "secret-password"
	passwordHash, err := utils.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	accessExpiresAt := time.Now().Add(15 * time.Minute)
	input := &authservice.LoginInput{
		Email:     email,
		Password:  password,
		UserAgent: "go-test",
		IPAddress: net.ParseIP("127.0.0.1"),
	}

	deps.identities.EXPECT().GetByEmail(ctx, email).Return(activeIdentity(identityID, email), nil)
	deps.credentials.EXPECT().GetByIdentityID(ctx, identityID).Return(&records.Credential{
		IdentityID:   identityID,
		PasswordHash: passwordHash,
	}, nil)
	deps.tokens.EXPECT().GenerateRefreshToken().Return("refresh-token", "refresh-hash", nil)
	deps.tokens.EXPECT().GenerateAccessToken(identityID, gomock.Any()).DoAndReturn(func(userID, sessionID uuid.UUID) (string, time.Time, error) {
		if userID != identityID {
			t.Fatalf("userID = %s, want %s", userID, identityID)
		}
		if sessionID == uuid.Nil {
			t.Fatal("sessionID is nil")
		}
		return "access-token", accessExpiresAt, nil
	})
	deps.sessions.EXPECT().Create(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, session *records.Session) error {
		if session.ID == uuid.Nil {
			t.Fatal("session.ID is nil")
		}
		if session.IdentityID != identityID {
			t.Fatalf("session.IdentityID = %s, want %s", session.IdentityID, identityID)
		}
		if session.RefreshTokenHash != "refresh-hash" {
			t.Fatalf("session.RefreshTokenHash = %q, want refresh-hash", session.RefreshTokenHash)
		}
		if session.UserAgent != input.UserAgent {
			t.Fatalf("session.UserAgent = %q, want %q", session.UserAgent, input.UserAgent)
		}
		if session.IPAddress == nil || !session.IPAddress.Equal(input.IPAddress) {
			t.Fatalf("session.IPAddress = %v, want %v", session.IPAddress, input.IPAddress)
		}
		if !session.ExpiresAt.After(session.CreatedAt) {
			t.Fatal("session.ExpiresAt must be after CreatedAt")
		}
		return nil
	})
	deps.authEvents.EXPECT().Create(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, event *records.AuthEvent) error {
		if event.IdentityID == nil || *event.IdentityID != identityID {
			t.Fatalf("event.IdentityID = %v, want %s", event.IdentityID, identityID)
		}
		if event.EventType != "login" {
			t.Fatalf("event.EventType = %q, want login", event.EventType)
		}
		if event.UserAgent != input.UserAgent {
			t.Fatalf("event.UserAgent = %q, want %q", event.UserAgent, input.UserAgent)
		}
		return nil
	})
	deps.events.EXPECT().PublishIdentityLogin(ctx, events.IdentityLoginPayload{
		IdentityID: identityID.String(),
		Email:      email,
	}).Return(nil)

	out, err := svc.Login(ctx, input)
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if out.AccessToken != "access-token" {
		t.Fatalf("Login().AccessToken = %q, want access-token", out.AccessToken)
	}
	if out.RefreshToken != "refresh-token" {
		t.Fatalf("Login().RefreshToken = %q, want refresh-token", out.RefreshToken)
	}
	if out.ExpiresIn <= 0 {
		t.Fatalf("Login().ExpiresIn = %d, want positive", out.ExpiresIn)
	}
}

func TestAuthService_Login_TransactionsNotConfigured(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, nil)

	_, err := svc.Login(context.Background(), &authservice.LoginInput{Email: "user@example.com", Password: "password"})
	if !errors.Is(err, authservice.ErrTransactionsNotConfigured) {
		t.Fatalf("Login() error = %v, want ErrTransactionsNotConfigured", err)
	}
}

func TestAuthService_Login_IdentityNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()

	deps.identities.EXPECT().GetByEmail(ctx, "user@example.com").Return(nil, pgx.ErrNoRows)

	_, err := svc.Login(ctx, &authservice.LoginInput{Email: "user@example.com", Password: "password"})
	if !errors.Is(err, authservice.ErrInvalidCredentials) {
		t.Fatalf("Login() error = %v, want ErrInvalidCredentials", err)
	}
}

func TestAuthService_Login_IdentityNotActive(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()
	identity := activeIdentity(uuid.New(), "user@example.com")
	identity.Status = "blocked"

	deps.identities.EXPECT().GetByEmail(ctx, identity.Email).Return(identity, nil)

	_, err := svc.Login(ctx, &authservice.LoginInput{Email: identity.Email, Password: "password"})
	if !errors.Is(err, authservice.ErrIdentityNotActive) {
		t.Fatalf("Login() error = %v, want ErrIdentityNotActive", err)
	}
}

func TestAuthService_Login_EmailNotVerified(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()
	identity := activeIdentity(uuid.New(), "user@example.com")
	identity.EmailVerified = false

	deps.identities.EXPECT().GetByEmail(ctx, identity.Email).Return(identity, nil)

	_, err := svc.Login(ctx, &authservice.LoginInput{Email: identity.Email, Password: "password"})
	if !errors.Is(err, authservice.ErrEmailNotVerified) {
		t.Fatalf("Login() error = %v, want ErrEmailNotVerified", err)
	}
}

func TestAuthService_Login_CredentialsNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()
	identity := activeIdentity(uuid.New(), "user@example.com")

	deps.identities.EXPECT().GetByEmail(ctx, identity.Email).Return(identity, nil)
	deps.credentials.EXPECT().GetByIdentityID(ctx, identity.ID).Return(nil, pgx.ErrNoRows)

	_, err := svc.Login(ctx, &authservice.LoginInput{Email: identity.Email, Password: "password"})
	if !errors.Is(err, authservice.ErrInvalidCredentials) {
		t.Fatalf("Login() error = %v, want ErrInvalidCredentials", err)
	}
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()
	identity := activeIdentity(uuid.New(), "user@example.com")
	passwordHash, err := utils.HashPassword("correct-password")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	deps.identities.EXPECT().GetByEmail(ctx, identity.Email).Return(identity, nil)
	deps.credentials.EXPECT().GetByIdentityID(ctx, identity.ID).Return(&records.Credential{IdentityID: identity.ID, PasswordHash: passwordHash}, nil)

	_, err = svc.Login(ctx, &authservice.LoginInput{Email: identity.Email, Password: "wrong-password"})
	if !errors.Is(err, authservice.ErrInvalidCredentials) {
		t.Fatalf("Login() error = %v, want ErrInvalidCredentials", err)
	}
}

func TestAuthService_Login_TokenError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()
	identity := activeIdentity(uuid.New(), "user@example.com")
	passwordHash, err := utils.HashPassword("password")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	tokenErr := errors.New("token error")

	deps.identities.EXPECT().GetByEmail(ctx, identity.Email).Return(identity, nil)
	deps.credentials.EXPECT().GetByIdentityID(ctx, identity.ID).Return(&records.Credential{IdentityID: identity.ID, PasswordHash: passwordHash}, nil)
	deps.tokens.EXPECT().GenerateRefreshToken().Return("", "", tokenErr)

	_, err = svc.Login(ctx, &authservice.LoginInput{Email: identity.Email, Password: "password"})
	if !errors.Is(err, tokenErr) {
		t.Fatalf("Login() error = %v, want tokenErr", err)
	}
}

func TestAuthService_Login_TxError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	txErr := errors.New("tx error")
	svc := newService(deps, func(context.Context, func(authservice.Repositories) error) error { return txErr })
	ctx := context.Background()
	identity := activeIdentity(uuid.New(), "user@example.com")
	passwordHash, err := utils.HashPassword("password")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	deps.identities.EXPECT().GetByEmail(ctx, identity.Email).Return(identity, nil)
	deps.credentials.EXPECT().GetByIdentityID(ctx, identity.ID).Return(&records.Credential{IdentityID: identity.ID, PasswordHash: passwordHash}, nil)
	deps.tokens.EXPECT().GenerateRefreshToken().Return("refresh-token", "refresh-hash", nil)
	deps.tokens.EXPECT().GenerateAccessToken(identity.ID, gomock.Any()).Return("access-token", time.Now().Add(time.Minute), nil)

	_, err = svc.Login(ctx, &authservice.LoginInput{Email: identity.Email, Password: "password"})
	if !errors.Is(err, txErr) {
		t.Fatalf("Login() error = %v, want txErr", err)
	}
}

func TestAuthService_Login_PublishError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()
	identity := activeIdentity(uuid.New(), "user@example.com")
	passwordHash, err := utils.HashPassword("password")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	publishErr := errors.New("publish error")

	deps.identities.EXPECT().GetByEmail(ctx, identity.Email).Return(identity, nil)
	deps.credentials.EXPECT().GetByIdentityID(ctx, identity.ID).Return(&records.Credential{IdentityID: identity.ID, PasswordHash: passwordHash}, nil)
	deps.tokens.EXPECT().GenerateRefreshToken().Return("refresh-token", "refresh-hash", nil)
	deps.tokens.EXPECT().GenerateAccessToken(identity.ID, gomock.Any()).Return("access-token", time.Now().Add(time.Minute), nil)
	deps.sessions.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.authEvents.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.events.EXPECT().PublishIdentityLogin(ctx, gomock.Any()).Return(publishErr)

	_, err = svc.Login(ctx, &authservice.LoginInput{Email: identity.Email, Password: "password"})
	if !errors.Is(err, publishErr) {
		t.Fatalf("Login() error = %v, want publishErr", err)
	}
}

func TestAuthService_Refresh_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()
	identityID := uuid.New()
	sessionID := uuid.New()
	refreshToken := "old-refresh-token"
	oldIP := net.ParseIP("127.0.0.1")
	session := &records.Session{
		ID:               sessionID,
		IdentityID:       identityID,
		RefreshTokenHash: utils.HashSHA256(refreshToken),
		UserAgent:        "go-test",
		IPAddress:        &oldIP,
		ExpiresAt:        time.Now().Add(time.Hour),
	}
	accessExpiresAt := time.Now().Add(15 * time.Minute)

	deps.sessions.EXPECT().GetByRefreshTokenHash(ctx, utils.HashSHA256(refreshToken)).Return(session, nil)
	deps.identities.EXPECT().GetByID(ctx, identityID).Return(activeIdentity(identityID, "user@example.com"), nil)
	deps.tokens.EXPECT().GenerateRefreshToken().Return("new-refresh-token", "new-refresh-hash", nil)
	deps.tokens.EXPECT().GenerateAccessToken(identityID, gomock.Any()).DoAndReturn(func(userID, newSessionID uuid.UUID) (string, time.Time, error) {
		if userID != identityID {
			t.Fatalf("userID = %s, want %s", userID, identityID)
		}
		if newSessionID == uuid.Nil || newSessionID == sessionID {
			t.Fatalf("newSessionID = %s, want new non-nil session", newSessionID)
		}
		return "new-access-token", accessExpiresAt, nil
	})
	deps.sessions.EXPECT().Revoke(ctx, sessionID).Return(nil)
	deps.sessions.EXPECT().Create(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, newSession *records.Session) error {
		if newSession.ID == uuid.Nil || newSession.ID == sessionID {
			t.Fatalf("newSession.ID = %s, want new non-nil session", newSession.ID)
		}
		if newSession.IdentityID != identityID {
			t.Fatalf("newSession.IdentityID = %s, want %s", newSession.IdentityID, identityID)
		}
		if newSession.RefreshTokenHash != "new-refresh-hash" {
			t.Fatalf("newSession.RefreshTokenHash = %q, want new-refresh-hash", newSession.RefreshTokenHash)
		}
		if newSession.UserAgent != session.UserAgent {
			t.Fatalf("newSession.UserAgent = %q, want %q", newSession.UserAgent, session.UserAgent)
		}
		return nil
	})
	deps.authEvents.EXPECT().Create(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, event *records.AuthEvent) error {
		if event.IdentityID == nil || *event.IdentityID != identityID {
			t.Fatalf("event.IdentityID = %v, want %s", event.IdentityID, identityID)
		}
		if event.EventType != "refresh" {
			t.Fatalf("event.EventType = %q, want refresh", event.EventType)
		}
		return nil
	})

	out, err := svc.Refresh(ctx, &authservice.RefreshInput{RefreshToken: refreshToken})
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}
	if out.AccessToken != "new-access-token" {
		t.Fatalf("Refresh().AccessToken = %q, want new-access-token", out.AccessToken)
	}
	if out.RefreshToken != "new-refresh-token" {
		t.Fatalf("Refresh().RefreshToken = %q, want new-refresh-token", out.RefreshToken)
	}
	if out.ExpiresIn <= 0 {
		t.Fatalf("Refresh().ExpiresIn = %d, want positive", out.ExpiresIn)
	}
}

func TestAuthService_Refresh_TransactionsNotConfigured(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, nil)

	_, err := svc.Refresh(context.Background(), &authservice.RefreshInput{RefreshToken: "refresh-token"})
	if !errors.Is(err, authservice.ErrTransactionsNotConfigured) {
		t.Fatalf("Refresh() error = %v, want ErrTransactionsNotConfigured", err)
	}
}

func TestAuthService_Refresh_SessionNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()

	deps.sessions.EXPECT().GetByRefreshTokenHash(ctx, utils.HashSHA256("refresh-token")).Return(nil, pgx.ErrNoRows)

	_, err := svc.Refresh(ctx, &authservice.RefreshInput{RefreshToken: "refresh-token"})
	if !errors.Is(err, authservice.ErrInvalidRefreshToken) {
		t.Fatalf("Refresh() error = %v, want ErrInvalidRefreshToken", err)
	}
}

func TestAuthService_Refresh_SessionRevoked(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()
	revokedAt := time.Now().UTC()
	session := &records.Session{ID: uuid.New(), IdentityID: uuid.New(), ExpiresAt: time.Now().Add(time.Hour), RevokedAt: &revokedAt}

	deps.sessions.EXPECT().GetByRefreshTokenHash(ctx, utils.HashSHA256("refresh-token")).Return(session, nil)

	_, err := svc.Refresh(ctx, &authservice.RefreshInput{RefreshToken: "refresh-token"})
	if !errors.Is(err, authservice.ErrSessionRevoked) {
		t.Fatalf("Refresh() error = %v, want ErrSessionRevoked", err)
	}
}

func TestAuthService_Refresh_SessionExpired(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()
	session := &records.Session{ID: uuid.New(), IdentityID: uuid.New(), ExpiresAt: time.Now().Add(-time.Minute)}

	deps.sessions.EXPECT().GetByRefreshTokenHash(ctx, utils.HashSHA256("refresh-token")).Return(session, nil)

	_, err := svc.Refresh(ctx, &authservice.RefreshInput{RefreshToken: "refresh-token"})
	if !errors.Is(err, authservice.ErrSessionExpired) {
		t.Fatalf("Refresh() error = %v, want ErrSessionExpired", err)
	}
}

func TestAuthService_Refresh_IdentityNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()
	identityID := uuid.New()
	session := &records.Session{ID: uuid.New(), IdentityID: identityID, ExpiresAt: time.Now().Add(time.Hour)}

	deps.sessions.EXPECT().GetByRefreshTokenHash(ctx, utils.HashSHA256("refresh-token")).Return(session, nil)
	deps.identities.EXPECT().GetByID(ctx, identityID).Return(nil, pgx.ErrNoRows)

	_, err := svc.Refresh(ctx, &authservice.RefreshInput{RefreshToken: "refresh-token"})
	if !errors.Is(err, authservice.ErrInvalidRefreshToken) {
		t.Fatalf("Refresh() error = %v, want ErrInvalidRefreshToken", err)
	}
}

func TestAuthService_Refresh_IdentityNotActive(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()
	identityID := uuid.New()
	identity := activeIdentity(identityID, "user@example.com")
	identity.Status = "blocked"
	session := &records.Session{ID: uuid.New(), IdentityID: identityID, ExpiresAt: time.Now().Add(time.Hour)}

	deps.sessions.EXPECT().GetByRefreshTokenHash(ctx, utils.HashSHA256("refresh-token")).Return(session, nil)
	deps.identities.EXPECT().GetByID(ctx, identityID).Return(identity, nil)

	_, err := svc.Refresh(ctx, &authservice.RefreshInput{RefreshToken: "refresh-token"})
	if !errors.Is(err, authservice.ErrIdentityNotActive) {
		t.Fatalf("Refresh() error = %v, want ErrIdentityNotActive", err)
	}
}

func TestAuthService_Refresh_TokenError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()
	identityID := uuid.New()
	tokenErr := errors.New("token error")
	session := &records.Session{ID: uuid.New(), IdentityID: identityID, ExpiresAt: time.Now().Add(time.Hour)}

	deps.sessions.EXPECT().GetByRefreshTokenHash(ctx, utils.HashSHA256("refresh-token")).Return(session, nil)
	deps.identities.EXPECT().GetByID(ctx, identityID).Return(activeIdentity(identityID, "user@example.com"), nil)
	deps.tokens.EXPECT().GenerateRefreshToken().Return("", "", tokenErr)

	_, err := svc.Refresh(ctx, &authservice.RefreshInput{RefreshToken: "refresh-token"})
	if !errors.Is(err, tokenErr) {
		t.Fatalf("Refresh() error = %v, want tokenErr", err)
	}
}

func TestAuthService_Refresh_TxError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	txErr := errors.New("tx error")
	svc := newService(deps, func(context.Context, func(authservice.Repositories) error) error { return txErr })
	ctx := context.Background()
	identityID := uuid.New()
	session := &records.Session{ID: uuid.New(), IdentityID: identityID, ExpiresAt: time.Now().Add(time.Hour)}

	deps.sessions.EXPECT().GetByRefreshTokenHash(ctx, utils.HashSHA256("refresh-token")).Return(session, nil)
	deps.identities.EXPECT().GetByID(ctx, identityID).Return(activeIdentity(identityID, "user@example.com"), nil)
	deps.tokens.EXPECT().GenerateRefreshToken().Return("new-refresh-token", "new-refresh-hash", nil)
	deps.tokens.EXPECT().GenerateAccessToken(identityID, gomock.Any()).Return("new-access-token", time.Now().Add(time.Minute), nil)

	_, err := svc.Refresh(ctx, &authservice.RefreshInput{RefreshToken: "refresh-token"})
	if !errors.Is(err, txErr) {
		t.Fatalf("Refresh() error = %v, want txErr", err)
	}
}

func TestAuthService_Logout_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	sessionID := uuid.New()
	ctx := authctx.WithAuth(context.Background(), identityID, sessionID)
	ip := net.ParseIP("127.0.0.1")
	session := &records.Session{ID: sessionID, IdentityID: identityID, UserAgent: "go-test", IPAddress: &ip, ExpiresAt: time.Now().Add(time.Hour)}

	deps.sessions.EXPECT().GetByID(ctx, sessionID).Return(session, nil)
	deps.sessions.EXPECT().Revoke(ctx, sessionID).Return(nil)
	deps.authEvents.EXPECT().Create(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, event *records.AuthEvent) error {
		if event.IdentityID == nil || *event.IdentityID != identityID {
			t.Fatalf("event.IdentityID = %v, want %s", event.IdentityID, identityID)
		}
		if event.EventType != "logout" {
			t.Fatalf("event.EventType = %q, want logout", event.EventType)
		}
		return nil
	})
	deps.events.EXPECT().PublishIdentityLogout(ctx, events.IdentityLogoutPayload{IdentityID: identityID.String()}).Return(nil)

	if err := svc.Logout(ctx); err != nil {
		t.Fatalf("Logout() error = %v", err)
	}
}

func TestAuthService_Logout_TransactionsNotConfigured(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, nil)
	ctx := authctx.WithAuth(context.Background(), uuid.New(), uuid.New())

	if err := svc.Logout(ctx); !errors.Is(err, authservice.ErrTransactionsNotConfigured) {
		t.Fatalf("Logout() error = %v, want ErrTransactionsNotConfigured", err)
	}
}

func TestAuthService_Logout_AuthContextMissing(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))

	if err := svc.Logout(context.Background()); !errors.Is(err, authctx.ErrAuthContextMissing) {
		t.Fatalf("Logout() error = %v, want ErrAuthContextMissing", err)
	}
}

func TestAuthService_Logout_SessionNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	sessionID := uuid.New()
	ctx := authctx.WithAuth(context.Background(), identityID, sessionID)

	deps.sessions.EXPECT().GetByID(ctx, sessionID).Return(nil, pgx.ErrNoRows)

	if err := svc.Logout(ctx); !errors.Is(err, authservice.ErrInvalidRefreshToken) {
		t.Fatalf("Logout() error = %v, want ErrInvalidRefreshToken", err)
	}
}

func TestAuthService_Logout_SessionIdentityMismatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	sessionID := uuid.New()
	ctx := authctx.WithAuth(context.Background(), identityID, sessionID)

	deps.sessions.EXPECT().GetByID(ctx, sessionID).Return(&records.Session{ID: sessionID, IdentityID: uuid.New()}, nil)

	if err := svc.Logout(ctx); !errors.Is(err, authservice.ErrInvalidRefreshToken) {
		t.Fatalf("Logout() error = %v, want ErrInvalidRefreshToken", err)
	}
}

func TestAuthService_Logout_SessionRevoked(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	sessionID := uuid.New()
	ctx := authctx.WithAuth(context.Background(), identityID, sessionID)
	revokedAt := time.Now().UTC()

	deps.sessions.EXPECT().GetByID(ctx, sessionID).Return(&records.Session{ID: sessionID, IdentityID: identityID, RevokedAt: &revokedAt}, nil)

	if err := svc.Logout(ctx); !errors.Is(err, authservice.ErrSessionRevoked) {
		t.Fatalf("Logout() error = %v, want ErrSessionRevoked", err)
	}
}

func TestAuthService_Logout_TxError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	txErr := errors.New("tx error")
	svc := newService(deps, func(context.Context, func(authservice.Repositories) error) error { return txErr })
	identityID := uuid.New()
	sessionID := uuid.New()
	ctx := authctx.WithAuth(context.Background(), identityID, sessionID)

	deps.sessions.EXPECT().GetByID(ctx, sessionID).Return(&records.Session{ID: sessionID, IdentityID: identityID}, nil)

	if err := svc.Logout(ctx); !errors.Is(err, txErr) {
		t.Fatalf("Logout() error = %v, want txErr", err)
	}
}

func TestAuthService_Logout_PublishError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	sessionID := uuid.New()
	ctx := authctx.WithAuth(context.Background(), identityID, sessionID)
	publishErr := errors.New("publish error")

	deps.sessions.EXPECT().GetByID(ctx, sessionID).Return(&records.Session{ID: sessionID, IdentityID: identityID}, nil)
	deps.sessions.EXPECT().Revoke(ctx, sessionID).Return(nil)
	deps.authEvents.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.events.EXPECT().PublishIdentityLogout(ctx, gomock.Any()).Return(publishErr)

	if err := svc.Logout(ctx); !errors.Is(err, publishErr) {
		t.Fatalf("Logout() error = %v, want publishErr", err)
	}
}
