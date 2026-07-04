package authservice_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/mock/gomock"

	"auth/internal/events"
	"auth/internal/models/records"
	authservice "auth/internal/services/auth"
	"auth/pkg/utils"
)

func TestAuthService_Register_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()
	input := &authservice.RegisterInput{Email: "user@example.com", Password: "secret-password"}

	deps.identities.EXPECT().GetByEmail(ctx, input.Email).Return(nil, pgx.ErrNoRows)
	deps.identities.EXPECT().Create(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, identity *records.Identity) error {
		if identity.ID == uuid.Nil {
			t.Fatal("identity.ID is nil")
		}
		if identity.Email != input.Email {
			t.Fatalf("identity.Email = %q, want %q", identity.Email, input.Email)
		}
		if identity.EmailVerified {
			t.Fatal("identity.EmailVerified = true, want false")
		}
		if identity.Status != "pending_verification" {
			t.Fatalf("identity.Status = %q, want pending_verification", identity.Status)
		}
		return nil
	})
	deps.credentials.EXPECT().Create(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, credential *records.Credential) error {
		if credential.IdentityID == uuid.Nil {
			t.Fatal("credential.IdentityID is nil")
		}
		if !utils.VerifyPassword(input.Password, credential.PasswordHash) {
			t.Fatal("credential.PasswordHash does not verify input password")
		}
		return nil
	})
	deps.verificationCodes.EXPECT().Create(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, code *records.VerificationCode) error {
		if code.ID == uuid.Nil {
			t.Fatal("code.ID is nil")
		}
		if code.IdentityID == nil || *code.IdentityID == uuid.Nil {
			t.Fatal("code.IdentityID is nil")
		}
		if code.Email == nil || *code.Email != input.Email {
			t.Fatalf("code.Email = %v, want %q", code.Email, input.Email)
		}
		if code.Purpose != "register" {
			t.Fatalf("code.Purpose = %q, want register", code.Purpose)
		}
		if code.CodeHash == "" {
			t.Fatal("code.CodeHash is empty")
		}
		if code.MaxAttempts != 5 {
			t.Fatalf("code.MaxAttempts = %d, want 5", code.MaxAttempts)
		}
		if !code.ExpiresAt.After(code.CreatedAt) {
			t.Fatal("code.ExpiresAt must be after CreatedAt")
		}
		return nil
	})
	deps.authEvents.EXPECT().Create(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, event *records.AuthEvent) error {
		if event.IdentityID == nil || *event.IdentityID == uuid.Nil {
			t.Fatal("event.IdentityID is nil")
		}
		if event.EventType != "register" {
			t.Fatalf("event.EventType = %q, want register", event.EventType)
		}
		return nil
	})
	deps.events.EXPECT().PublishIdentityCreated(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, payload events.IdentityCreatedPayload) error {
		if payload.IdentityID == "" {
			t.Fatal("payload.IdentityID is empty")
		}
		if payload.Email != input.Email {
			t.Fatalf("payload.Email = %q, want %q", payload.Email, input.Email)
		}
		return nil
	})
	deps.events.EXPECT().PublishVerificationCodeSend(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, payload events.VerificationCodeSendPayload) error {
		if payload.IdentityID == nil || *payload.IdentityID == "" {
			t.Fatal("payload.IdentityID is nil or empty")
		}
		if payload.Email != input.Email {
			t.Fatalf("payload.Email = %q, want %q", payload.Email, input.Email)
		}
		if len(payload.Code) != 6 {
			t.Fatalf("len(payload.Code) = %d, want 6", len(payload.Code))
		}
		if payload.Purpose != "register" {
			t.Fatalf("payload.Purpose = %q, want register", payload.Purpose)
		}
		return nil
	})

	out, err := svc.Register(ctx, input)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if out.IdentityID == "" {
		t.Fatal("Register().IdentityID is empty")
	}
	if out.Email != input.Email {
		t.Fatalf("Register().Email = %q, want %q", out.Email, input.Email)
	}
	if out.Status != "pending_verification" {
		t.Fatalf("Register().Status = %q, want pending_verification", out.Status)
	}
}

func TestAuthService_Register_TransactionsNotConfigured(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, nil)

	_, err := svc.Register(context.Background(), &authservice.RegisterInput{Email: "user@example.com", Password: "password"})
	if !errors.Is(err, authservice.ErrTransactionsNotConfigured) {
		t.Fatalf("Register() error = %v, want ErrTransactionsNotConfigured", err)
	}
}

func TestAuthService_Register_EmailAlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()
	email := "user@example.com"

	deps.identities.EXPECT().GetByEmail(ctx, email).Return(&records.Identity{ID: uuid.New(), Email: email}, nil)

	_, err := svc.Register(ctx, &authservice.RegisterInput{Email: email, Password: "password"})
	if !errors.Is(err, authservice.ErrEmailAlreadyExists) {
		t.Fatalf("Register() error = %v, want ErrEmailAlreadyExists", err)
	}
}

func TestAuthService_Register_TxError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()
	email := "user@example.com"
	wantErr := errors.New("create identity")

	deps.identities.EXPECT().GetByEmail(ctx, email).Return(nil, pgx.ErrNoRows)
	deps.identities.EXPECT().Create(ctx, gomock.Any()).Return(wantErr)

	_, err := svc.Register(ctx, &authservice.RegisterInput{Email: email, Password: "password"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("Register() error = %v, want %v", err, wantErr)
	}
}

func TestAuthService_Register_PublishError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()
	email := "user@example.com"
	wantErr := errors.New("publish identity created")

	deps.identities.EXPECT().GetByEmail(ctx, email).Return(nil, pgx.ErrNoRows)
	deps.identities.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.credentials.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.verificationCodes.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.authEvents.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.events.EXPECT().PublishIdentityCreated(ctx, gomock.Any()).Return(wantErr)

	_, err := svc.Register(ctx, &authservice.RegisterInput{Email: email, Password: "password"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("Register() error = %v, want %v", err, wantErr)
	}
}

func TestAuthService_RegisterVerify_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()
	identityID := uuid.New()
	codeID := uuid.New()
	email := "user@example.com"
	verificationCode := "123456"

	deps.verificationCodes.EXPECT().GetActiveByEmailAndPurpose(ctx, email, "register").Return(&records.VerificationCode{
		ID:            codeID,
		IdentityID:    &identityID,
		Email:         &email,
		Purpose:       "register",
		CodeHash:      utils.HashSHA256(verificationCode),
		AttemptsCount: 0,
		MaxAttempts:   5,
		ExpiresAt:     time.Now().Add(time.Minute),
		CreatedAt:     time.Now(),
	}, nil)
	deps.identities.EXPECT().GetByID(ctx, identityID).Return(&records.Identity{ID: identityID, Email: email, EmailVerified: false, Status: "pending_verification"}, nil)
	deps.identities.EXPECT().SetEmailVerified(ctx, identityID).Return(nil)
	deps.identities.EXPECT().SetStatus(ctx, identityID, "active").Return(nil)
	deps.verificationCodes.EXPECT().Consume(ctx, codeID).Return(nil)
	deps.authEvents.EXPECT().Create(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, event *records.AuthEvent) error {
		if event.IdentityID == nil || *event.IdentityID != identityID {
			t.Fatalf("event.IdentityID = %v, want %s", event.IdentityID, identityID)
		}
		if event.EventType != "register_verified" {
			t.Fatalf("event.EventType = %q, want register_verified", event.EventType)
		}
		return nil
	})
	deps.events.EXPECT().PublishIdentityUpdated(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, payload events.IdentityUpdatedPayload) error {
		if payload.IdentityID != identityID.String() {
			t.Fatalf("payload.IdentityID = %q, want %q", payload.IdentityID, identityID.String())
		}
		if payload.Email != email {
			t.Fatalf("payload.Email = %q, want %q", payload.Email, email)
		}
		if payload.Status != "active" {
			t.Fatalf("payload.Status = %q, want active", payload.Status)
		}
		if payload.EmailVerified == nil || !*payload.EmailVerified {
			t.Fatal("payload.EmailVerified is nil or false")
		}
		return nil
	})
	message, err := svc.RegisterVerify(ctx, &authservice.VerifyCodeInput{Email: email, Code: verificationCode})
	if err != nil {
		t.Fatalf("RegisterVerify() error = %v", err)
	}
	if message != "registration verified" {
		t.Fatalf("RegisterVerify() message = %q, want registration verified", message)
	}
}

func TestAuthService_RegisterVerify_CodeNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()

	deps.verificationCodes.EXPECT().GetActiveByEmailAndPurpose(ctx, "user@example.com", "register").Return(nil, pgx.ErrNoRows)

	_, err := svc.RegisterVerify(ctx, &authservice.VerifyCodeInput{Email: "user@example.com", Code: "123456"})
	if !errors.Is(err, authservice.ErrInvalidCode) {
		t.Fatalf("RegisterVerify() error = %v, want ErrInvalidCode", err)
	}
}

func TestAuthService_RegisterVerify_TooManyAttempts(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()
	identityID := uuid.New()
	email := "user@example.com"

	deps.verificationCodes.EXPECT().GetActiveByEmailAndPurpose(ctx, email, "register").Return(&records.VerificationCode{
		ID:            uuid.New(),
		IdentityID:    &identityID,
		Email:         &email,
		Purpose:       "register",
		CodeHash:      utils.HashSHA256("123456"),
		AttemptsCount: 5,
		MaxAttempts:   5,
	}, nil)

	_, err := svc.RegisterVerify(ctx, &authservice.VerifyCodeInput{Email: email, Code: "123456"})
	if !errors.Is(err, authservice.ErrTooManyAttempts) {
		t.Fatalf("RegisterVerify() error = %v, want ErrTooManyAttempts", err)
	}
}

func TestAuthService_RegisterVerify_InvalidCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()
	identityID := uuid.New()
	codeID := uuid.New()
	email := "user@example.com"

	deps.verificationCodes.EXPECT().GetActiveByEmailAndPurpose(ctx, email, "register").Return(&records.VerificationCode{
		ID:            codeID,
		IdentityID:    &identityID,
		Email:         &email,
		Purpose:       "register",
		CodeHash:      utils.HashSHA256("123456"),
		AttemptsCount: 0,
		MaxAttempts:   5,
	}, nil)
	deps.verificationCodes.EXPECT().IncrementAttempts(ctx, codeID).Return(nil)

	_, err := svc.RegisterVerify(ctx, &authservice.VerifyCodeInput{Email: email, Code: "000000"})
	if !errors.Is(err, authservice.ErrInvalidCode) {
		t.Fatalf("RegisterVerify() error = %v, want ErrInvalidCode", err)
	}
}

func TestAuthService_RegisterVerify_TxError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()
	identityID := uuid.New()
	email := "user@example.com"
	wantErr := errors.New("set verified")

	deps.verificationCodes.EXPECT().GetActiveByEmailAndPurpose(ctx, email, "register").Return(&records.VerificationCode{
		ID:            uuid.New(),
		IdentityID:    &identityID,
		Email:         &email,
		Purpose:       "register",
		CodeHash:      utils.HashSHA256("123456"),
		AttemptsCount: 0,
		MaxAttempts:   5,
	}, nil)
	deps.identities.EXPECT().GetByID(ctx, identityID).Return(&records.Identity{ID: identityID, Email: email, EmailVerified: false, Status: "pending_verification"}, nil)
	deps.identities.EXPECT().SetEmailVerified(ctx, identityID).Return(wantErr)

	_, err := svc.RegisterVerify(ctx, &authservice.VerifyCodeInput{Email: email, Code: "123456"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("RegisterVerify() error = %v, want %v", err, wantErr)
	}
}

func TestAuthService_ResendVerificationCode_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()
	identityID := uuid.New()
	email := "user@example.com"
	identity := &records.Identity{ID: identityID, Email: email, EmailVerified: false, Status: "pending_verification"}

	deps.identities.EXPECT().GetByEmail(ctx, email).Return(identity, nil)
	deps.rateLimiter.EXPECT().Allow(ctx, "rate:verification_code:register:"+email, time.Minute, 15*time.Minute, int64(5)).Return(nil)
	deps.verificationCodes.EXPECT().Create(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, code *records.VerificationCode) error {
		if code.IdentityID == nil || *code.IdentityID != identityID {
			t.Fatalf("code.IdentityID = %v, want %s", code.IdentityID, identityID)
		}
		if code.Email == nil || *code.Email != email {
			t.Fatalf("code.Email = %v, want %q", code.Email, email)
		}
		if code.Purpose != "register" {
			t.Fatalf("code.Purpose = %q, want register", code.Purpose)
		}
		if code.CodeHash == "" {
			t.Fatal("code.CodeHash is empty")
		}
		return nil
	})
	deps.authEvents.EXPECT().Create(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, event *records.AuthEvent) error {
		if event.IdentityID == nil || *event.IdentityID != identityID {
			t.Fatalf("event.IdentityID = %v, want %s", event.IdentityID, identityID)
		}
		if event.EventType != "register_code_resent" {
			t.Fatalf("event.EventType = %q, want register_code_resent", event.EventType)
		}
		return nil
	})
	deps.events.EXPECT().PublishVerificationCodeSend(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, payload events.VerificationCodeSendPayload) error {
		if payload.IdentityID == nil || *payload.IdentityID != identityID.String() {
			t.Fatalf("payload.IdentityID = %v, want %s", payload.IdentityID, identityID)
		}
		if payload.Email != email {
			t.Fatalf("payload.Email = %q, want %q", payload.Email, email)
		}
		if len(payload.Code) != 6 {
			t.Fatalf("len(payload.Code) = %d, want 6", len(payload.Code))
		}
		if payload.Purpose != "register" {
			t.Fatalf("payload.Purpose = %q, want register", payload.Purpose)
		}
		return nil
	})

	message, err := svc.ResendVerificationCode(ctx, &authservice.ResendCodeInput{Email: email})
	if err != nil {
		t.Fatalf("ResendVerificationCode() error = %v", err)
	}
	if message != "verification code sent" {
		t.Fatalf("ResendVerificationCode() message = %q, want verification code sent", message)
	}
}

func TestAuthService_ResendVerificationCode_RateLimiterNotConfigured(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newServiceWithoutRateLimiter(deps, transactionWithMocks(deps))

	_, err := svc.ResendVerificationCode(context.Background(), &authservice.ResendCodeInput{Email: "user@example.com"})
	if !errors.Is(err, authservice.ErrRateLimiterNotConfigured) {
		t.Fatalf("ResendVerificationCode() error = %v, want ErrRateLimiterNotConfigured", err)
	}
}

func TestAuthService_ResendVerificationCode_IdentityNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()
	email := "user@example.com"

	deps.identities.EXPECT().GetByEmail(ctx, email).Return(nil, pgx.ErrNoRows)

	_, err := svc.ResendVerificationCode(ctx, &authservice.ResendCodeInput{Email: email})
	if !errors.Is(err, authservice.ErrIdentityNotFound) {
		t.Fatalf("ResendVerificationCode() error = %v, want ErrIdentityNotFound", err)
	}
}

func TestAuthService_ResendVerificationCode_EmailAlreadyVerified(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()
	email := "user@example.com"

	deps.identities.EXPECT().GetByEmail(ctx, email).Return(&records.Identity{ID: uuid.New(), Email: email, EmailVerified: true, Status: "active"}, nil)

	_, err := svc.ResendVerificationCode(ctx, &authservice.ResendCodeInput{Email: email})
	if !errors.Is(err, authservice.ErrEmailAlreadyVerified) {
		t.Fatalf("ResendVerificationCode() error = %v, want ErrEmailAlreadyVerified", err)
	}
}

func TestAuthService_ResendVerificationCode_RateLimited(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()
	email := "user@example.com"
	wantErr := errors.New("rate limited")

	deps.identities.EXPECT().GetByEmail(ctx, email).Return(&records.Identity{ID: uuid.New(), Email: email, Status: "pending_verification"}, nil)
	deps.rateLimiter.EXPECT().Allow(ctx, "rate:verification_code:register:"+email, time.Minute, 15*time.Minute, int64(5)).Return(wantErr)

	_, err := svc.ResendVerificationCode(ctx, &authservice.ResendCodeInput{Email: email})
	if !errors.Is(err, wantErr) {
		t.Fatalf("ResendVerificationCode() error = %v, want %v", err, wantErr)
	}
}

func TestAuthService_ResendVerificationCode_TxError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()
	email := "user@example.com"
	wantErr := errors.New("create code")

	deps.identities.EXPECT().GetByEmail(ctx, email).Return(&records.Identity{ID: uuid.New(), Email: email, Status: "pending_verification"}, nil)
	deps.rateLimiter.EXPECT().Allow(ctx, "rate:verification_code:register:"+email, time.Minute, 15*time.Minute, int64(5)).Return(nil)
	deps.verificationCodes.EXPECT().Create(ctx, gomock.Any()).Return(wantErr)

	_, err := svc.ResendVerificationCode(ctx, &authservice.ResendCodeInput{Email: email})
	if !errors.Is(err, wantErr) {
		t.Fatalf("ResendVerificationCode() error = %v, want %v", err, wantErr)
	}
}

func TestAuthService_ResendVerificationCode_PublishError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()
	email := "user@example.com"
	wantErr := errors.New("publish code")

	deps.identities.EXPECT().GetByEmail(ctx, email).Return(&records.Identity{ID: uuid.New(), Email: email, Status: "pending_verification"}, nil)
	deps.rateLimiter.EXPECT().Allow(ctx, "rate:verification_code:register:"+email, time.Minute, 15*time.Minute, int64(5)).Return(nil)
	deps.verificationCodes.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.authEvents.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.events.EXPECT().PublishVerificationCodeSend(ctx, gomock.Any()).Return(wantErr)

	_, err := svc.ResendVerificationCode(ctx, &authservice.ResendCodeInput{Email: email})
	if !errors.Is(err, wantErr) {
		t.Fatalf("ResendVerificationCode() error = %v, want %v", err, wantErr)
	}
}
