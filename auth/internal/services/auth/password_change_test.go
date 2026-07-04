package authservice_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/mock/gomock"

	"auth/internal/authctx"
	"auth/internal/models/records"
	authservice "auth/internal/services/auth"
	"auth/pkg/utils"
)

func authCtx(identityID uuid.UUID) context.Context {
	return authctx.WithAuth(context.Background(), identityID, uuid.New())
}

func TestAuthService_ChangePassword_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)
	email := "user@example.com"
	currentPassword := "old-password"
	passwordHash, _ := utils.HashPassword(currentPassword)

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(activeIdentity(identityID, email), nil)
	deps.credentials.EXPECT().GetByIdentityID(ctx, identityID).Return(&records.Credential{IdentityID: identityID, PasswordHash: passwordHash}, nil)
	deps.rateLimiter.EXPECT().Allow(ctx, "rate:verification_code:password_change:"+identityID.String(), time.Minute, 15*time.Minute, int64(5)).Return(nil)
	deps.verificationCodes.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.authEvents.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.events.EXPECT().PublishPasswordChangeCodeSend(ctx, gomock.Any()).Return(nil)

	msg, err := svc.ChangePassword(ctx, &authservice.ChangePasswordInput{CurrentPassword: currentPassword})
	if err != nil {
		t.Fatalf("ChangePassword() error = %v", err)
	}
	if msg != "password change code sent" {
		t.Fatalf("ChangePassword() message = %q, want password change code sent", msg)
	}
}

func TestAuthService_ChangePassword_TransactionsNotConfigured(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, nil)
	ctx := authCtx(uuid.New())

	_, err := svc.ChangePassword(ctx, &authservice.ChangePasswordInput{CurrentPassword: "password"})
	if !errors.Is(err, authservice.ErrTransactionsNotConfigured) {
		t.Fatalf("ChangePassword() error = %v, want ErrTransactionsNotConfigured", err)
	}
}

func TestAuthService_ChangePassword_RateLimiterNotConfigured(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newServiceWithoutRateLimiter(deps, transactionWithMocks(deps))
	ctx := authCtx(uuid.New())

	_, err := svc.ChangePassword(ctx, &authservice.ChangePasswordInput{CurrentPassword: "password"})
	if !errors.Is(err, authservice.ErrRateLimiterNotConfigured) {
		t.Fatalf("ChangePassword() error = %v, want ErrRateLimiterNotConfigured", err)
	}
}

func TestAuthService_ChangePassword_IdentityNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(nil, pgx.ErrNoRows)

	_, err := svc.ChangePassword(ctx, &authservice.ChangePasswordInput{CurrentPassword: "password"})
	if err == nil {
		t.Fatal("ChangePassword() expected error, got nil")
	}
}

func TestAuthService_ChangePassword_IdentityNotActive(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)
	identity := activeIdentity(identityID, "user@example.com")
	identity.Status = "blocked"

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(identity, nil)

	_, err := svc.ChangePassword(ctx, &authservice.ChangePasswordInput{CurrentPassword: "password"})
	if !errors.Is(err, authservice.ErrIdentityNotActive) {
		t.Fatalf("ChangePassword() error = %v, want ErrIdentityNotActive", err)
	}
}

func TestAuthService_ChangePassword_EmailNotVerified(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)
	identity := activeIdentity(identityID, "user@example.com")
	identity.EmailVerified = false

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(identity, nil)

	_, err := svc.ChangePassword(ctx, &authservice.ChangePasswordInput{CurrentPassword: "password"})
	if !errors.Is(err, authservice.ErrEmailNotVerified) {
		t.Fatalf("ChangePassword() error = %v, want ErrEmailNotVerified", err)
	}
}

func TestAuthService_ChangePassword_CurrentPasswordIncorrect(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)
	passwordHash, _ := utils.HashPassword("correct-password")

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(activeIdentity(identityID, "user@example.com"), nil)
	deps.credentials.EXPECT().GetByIdentityID(ctx, identityID).Return(&records.Credential{IdentityID: identityID, PasswordHash: passwordHash}, nil)

	_, err := svc.ChangePassword(ctx, &authservice.ChangePasswordInput{CurrentPassword: "wrong-password"})
	if !errors.Is(err, authservice.ErrCurrentPasswordIncorrect) {
		t.Fatalf("ChangePassword() error = %v, want ErrCurrentPasswordIncorrect", err)
	}
}

func TestAuthService_ChangePassword_RateLimited(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)
	passwordHash, _ := utils.HashPassword("password")
	rateErr := errors.New("rate limited")

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(activeIdentity(identityID, "user@example.com"), nil)
	deps.credentials.EXPECT().GetByIdentityID(ctx, identityID).Return(&records.Credential{IdentityID: identityID, PasswordHash: passwordHash}, nil)
	deps.rateLimiter.EXPECT().Allow(ctx, "rate:verification_code:password_change:"+identityID.String(), time.Minute, 15*time.Minute, int64(5)).Return(rateErr)

	_, err := svc.ChangePassword(ctx, &authservice.ChangePasswordInput{CurrentPassword: "password"})
	if !errors.Is(err, rateErr) {
		t.Fatalf("ChangePassword() error = %v, want rateErr", err)
	}
}

func TestAuthService_ChangePassword_TxError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)
	passwordHash, _ := utils.HashPassword("password")
	txErr := errors.New("tx error")

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(activeIdentity(identityID, "user@example.com"), nil)
	deps.credentials.EXPECT().GetByIdentityID(ctx, identityID).Return(&records.Credential{IdentityID: identityID, PasswordHash: passwordHash}, nil)
	deps.rateLimiter.EXPECT().Allow(ctx, "rate:verification_code:password_change:"+identityID.String(), time.Minute, 15*time.Minute, int64(5)).Return(nil)
	deps.verificationCodes.EXPECT().Create(ctx, gomock.Any()).Return(txErr)

	_, err := svc.ChangePassword(ctx, &authservice.ChangePasswordInput{CurrentPassword: "password"})
	if !errors.Is(err, txErr) {
		t.Fatalf("ChangePassword() error = %v, want txErr", err)
	}
}

func TestAuthService_ChangePassword_PublishError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)
	passwordHash, _ := utils.HashPassword("password")
	publishErr := errors.New("publish error")

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(activeIdentity(identityID, "user@example.com"), nil)
	deps.credentials.EXPECT().GetByIdentityID(ctx, identityID).Return(&records.Credential{IdentityID: identityID, PasswordHash: passwordHash}, nil)
	deps.rateLimiter.EXPECT().Allow(ctx, "rate:verification_code:password_change:"+identityID.String(), time.Minute, 15*time.Minute, int64(5)).Return(nil)
	deps.verificationCodes.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.authEvents.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.events.EXPECT().PublishPasswordChangeCodeSend(ctx, gomock.Any()).Return(publishErr)

	_, err := svc.ChangePassword(ctx, &authservice.ChangePasswordInput{CurrentPassword: "password"})
	if !errors.Is(err, publishErr) {
		t.Fatalf("ChangePassword() error = %v, want publishErr", err)
	}
}

func TestAuthService_ChangePasswordVerify_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	codeID := uuid.New()
	ctx := authCtx(identityID)
	code := "123456"

	deps.verificationCodes.EXPECT().GetActiveByIdentityIDAndPurpose(ctx, identityID, "password_change").Return(&records.VerificationCode{
		ID:            codeID,
		IdentityID:    &identityID,
		Purpose:       "password_change",
		CodeHash:      utils.HashSHA256(code),
		AttemptsCount: 0,
		MaxAttempts:   5,
	}, nil)
	deps.tokens.EXPECT().GenerateRefreshToken().Return("change-token", "change-token-hash", nil)
	deps.passwordChangeTokens.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.verificationCodes.EXPECT().Consume(ctx, codeID).Return(nil)
	deps.authEvents.EXPECT().Create(ctx, gomock.Any()).Return(nil)

	out, err := svc.ChangePasswordVerify(ctx, &authservice.ChangePasswordVerifyInput{Code: code})
	if err != nil {
		t.Fatalf("ChangePasswordVerify() error = %v", err)
	}
	if out.ChangeToken != "change-token" {
		t.Fatalf("ChangePasswordVerify().ChangeToken = %q, want change-token", out.ChangeToken)
	}
	if out.ExpiresIn <= 0 {
		t.Fatalf("ChangePasswordVerify().ExpiresIn = %d, want positive", out.ExpiresIn)
	}
}

func TestAuthService_ChangePasswordVerify_TransactionsNotConfigured(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, nil)
	ctx := authCtx(uuid.New())

	_, err := svc.ChangePasswordVerify(ctx, &authservice.ChangePasswordVerifyInput{Code: "123456"})
	if !errors.Is(err, authservice.ErrTransactionsNotConfigured) {
		t.Fatalf("ChangePasswordVerify() error = %v, want ErrTransactionsNotConfigured", err)
	}
}

func TestAuthService_ChangePasswordVerify_CodeNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)

	deps.verificationCodes.EXPECT().GetActiveByIdentityIDAndPurpose(ctx, identityID, "password_change").Return(nil, pgx.ErrNoRows)

	_, err := svc.ChangePasswordVerify(ctx, &authservice.ChangePasswordVerifyInput{Code: "123456"})
	if !errors.Is(err, authservice.ErrInvalidCode) {
		t.Fatalf("ChangePasswordVerify() error = %v, want ErrInvalidCode", err)
	}
}

func TestAuthService_ChangePasswordVerify_TooManyAttempts(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)

	deps.verificationCodes.EXPECT().GetActiveByIdentityIDAndPurpose(ctx, identityID, "password_change").Return(&records.VerificationCode{
		ID:            uuid.New(),
		IdentityID:    &identityID,
		Purpose:       "password_change",
		CodeHash:      utils.HashSHA256("123456"),
		AttemptsCount: 5,
		MaxAttempts:   5,
	}, nil)

	_, err := svc.ChangePasswordVerify(ctx, &authservice.ChangePasswordVerifyInput{Code: "123456"})
	if !errors.Is(err, authservice.ErrTooManyAttempts) {
		t.Fatalf("ChangePasswordVerify() error = %v, want ErrTooManyAttempts", err)
	}
}

func TestAuthService_ChangePasswordVerify_InvalidCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	codeID := uuid.New()
	ctx := authCtx(identityID)

	deps.verificationCodes.EXPECT().GetActiveByIdentityIDAndPurpose(ctx, identityID, "password_change").Return(&records.VerificationCode{
		ID:            codeID,
		IdentityID:    &identityID,
		Purpose:       "password_change",
		CodeHash:      utils.HashSHA256("123456"),
		AttemptsCount: 0,
		MaxAttempts:   5,
	}, nil)
	deps.verificationCodes.EXPECT().IncrementAttempts(ctx, codeID).Return(nil)

	_, err := svc.ChangePasswordVerify(ctx, &authservice.ChangePasswordVerifyInput{Code: "000000"})
	if !errors.Is(err, authservice.ErrInvalidCode) {
		t.Fatalf("ChangePasswordVerify() error = %v, want ErrInvalidCode", err)
	}
}

func TestAuthService_ChangePasswordVerify_TxError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)
	txErr := errors.New("tx error")

	deps.verificationCodes.EXPECT().GetActiveByIdentityIDAndPurpose(ctx, identityID, "password_change").Return(&records.VerificationCode{
		ID:            uuid.New(),
		IdentityID:    &identityID,
		Purpose:       "password_change",
		CodeHash:      utils.HashSHA256("123456"),
		AttemptsCount: 0,
		MaxAttempts:   5,
	}, nil)
	deps.tokens.EXPECT().GenerateRefreshToken().Return("change-token", "change-token-hash", nil)
	deps.passwordChangeTokens.EXPECT().Create(ctx, gomock.Any()).Return(txErr)

	_, err := svc.ChangePasswordVerify(ctx, &authservice.ChangePasswordVerifyInput{Code: "123456"})
	if !errors.Is(err, txErr) {
		t.Fatalf("ChangePasswordVerify() error = %v, want txErr", err)
	}
}

func TestAuthService_CompletePasswordChange_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)
	now := time.Now().UTC()

	deps.passwordChangeTokens.EXPECT().GetByTokenHash(ctx, utils.HashSHA256("change-token")).Return(&records.PasswordChangeToken{
		ID:         uuid.New(),
		IdentityID: identityID,
		TokenHash:  "change-token-hash",
		ExpiresAt:  now.Add(time.Hour),
		CreatedAt:  now,
	}, nil)
	deps.identities.EXPECT().GetByID(ctx, identityID).Return(activeIdentity(identityID, "user@example.com"), nil)
	deps.credentials.EXPECT().UpdatePassword(ctx, identityID, gomock.Any()).Return(nil)
	deps.passwordChangeTokens.EXPECT().Consume(ctx, gomock.Any()).Return(nil)
	deps.sessions.EXPECT().RevokeAllByIdentityID(ctx, identityID).Return(nil)
	deps.authEvents.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.events.EXPECT().PublishIdentityUpdated(ctx, gomock.Any()).Return(nil)

	msg, err := svc.CompletePasswordChange(ctx, &authservice.CompletePasswordChangeInput{ChangeToken: "change-token", NewPassword: "new-password"})
	if err != nil {
		t.Fatalf("CompletePasswordChange() error = %v", err)
	}
	if msg != "password changed" {
		t.Fatalf("CompletePasswordChange() message = %q, want password changed", msg)
	}
}

func TestAuthService_CompletePasswordChange_TransactionsNotConfigured(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, nil)
	ctx := authCtx(uuid.New())

	_, err := svc.CompletePasswordChange(ctx, &authservice.CompletePasswordChangeInput{ChangeToken: "change-token", NewPassword: "new-password"})
	if !errors.Is(err, authservice.ErrTransactionsNotConfigured) {
		t.Fatalf("CompletePasswordChange() error = %v, want ErrTransactionsNotConfigured", err)
	}
}

func TestAuthService_CompletePasswordChange_TokenNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)

	deps.passwordChangeTokens.EXPECT().GetByTokenHash(ctx, utils.HashSHA256("change-token")).Return(nil, pgx.ErrNoRows)

	_, err := svc.CompletePasswordChange(ctx, &authservice.CompletePasswordChangeInput{ChangeToken: "change-token", NewPassword: "new-password"})
	if !errors.Is(err, authservice.ErrInvalidChangeToken) {
		t.Fatalf("CompletePasswordChange() error = %v, want ErrInvalidChangeToken", err)
	}
}

func TestAuthService_CompletePasswordChange_TokenWrongIdentity(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)
	now := time.Now().UTC()

	deps.passwordChangeTokens.EXPECT().GetByTokenHash(ctx, utils.HashSHA256("change-token")).Return(&records.PasswordChangeToken{
		ID:         uuid.New(),
		IdentityID: uuid.New(),
		TokenHash:  "change-token-hash",
		ExpiresAt:  now.Add(time.Hour),
		CreatedAt:  now,
	}, nil)

	_, err := svc.CompletePasswordChange(ctx, &authservice.CompletePasswordChangeInput{ChangeToken: "change-token", NewPassword: "new-password"})
	if !errors.Is(err, authservice.ErrInvalidChangeToken) {
		t.Fatalf("CompletePasswordChange() error = %v, want ErrInvalidChangeToken", err)
	}
}

func TestAuthService_CompletePasswordChange_TxError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)
	now := time.Now().UTC()
	txErr := errors.New("tx error")

	deps.passwordChangeTokens.EXPECT().GetByTokenHash(ctx, utils.HashSHA256("change-token")).Return(&records.PasswordChangeToken{
		ID:         uuid.New(),
		IdentityID: identityID,
		TokenHash:  "change-token-hash",
		ExpiresAt:  now.Add(time.Hour),
		CreatedAt:  now,
	}, nil)
	deps.identities.EXPECT().GetByID(ctx, identityID).Return(activeIdentity(identityID, "user@example.com"), nil)
	deps.credentials.EXPECT().UpdatePassword(ctx, identityID, gomock.Any()).Return(txErr)

	_, err := svc.CompletePasswordChange(ctx, &authservice.CompletePasswordChangeInput{ChangeToken: "change-token", NewPassword: "new-password"})
	if !errors.Is(err, txErr) {
		t.Fatalf("CompletePasswordChange() error = %v, want txErr", err)
	}
}

func TestAuthService_CompletePasswordChange_PublishError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)
	now := time.Now().UTC()
	publishErr := errors.New("publish error")

	deps.passwordChangeTokens.EXPECT().GetByTokenHash(ctx, utils.HashSHA256("change-token")).Return(&records.PasswordChangeToken{
		ID:         uuid.New(),
		IdentityID: identityID,
		TokenHash:  "change-token-hash",
		ExpiresAt:  now.Add(time.Hour),
		CreatedAt:  now,
	}, nil)
	deps.identities.EXPECT().GetByID(ctx, identityID).Return(activeIdentity(identityID, "user@example.com"), nil)
	deps.credentials.EXPECT().UpdatePassword(ctx, identityID, gomock.Any()).Return(nil)
	deps.passwordChangeTokens.EXPECT().Consume(ctx, gomock.Any()).Return(nil)
	deps.sessions.EXPECT().RevokeAllByIdentityID(ctx, identityID).Return(nil)
	deps.authEvents.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.events.EXPECT().PublishIdentityUpdated(ctx, gomock.Any()).Return(publishErr)

	_, err := svc.CompletePasswordChange(ctx, &authservice.CompletePasswordChangeInput{ChangeToken: "change-token", NewPassword: "new-password"})
	if !errors.Is(err, publishErr) {
		t.Fatalf("CompletePasswordChange() error = %v, want publishErr", err)
	}
}

func TestAuthService_ChangePasswordCodeResend_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)
	email := "user@example.com"

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(activeIdentity(identityID, email), nil)
	deps.verificationCodes.EXPECT().GetActiveByIdentityIDAndPurpose(ctx, identityID, "password_change").Return(&records.VerificationCode{
		ID:         uuid.New(),
		IdentityID: &identityID,
		Purpose:    "password_change",
	}, nil)
	deps.rateLimiter.EXPECT().Allow(ctx, "rate:verification_code:password_change:"+identityID.String(), time.Minute, 15*time.Minute, int64(5)).Return(nil)
	deps.verificationCodes.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.authEvents.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.events.EXPECT().PublishPasswordChangeCodeSend(ctx, gomock.Any()).Return(nil)

	msg, err := svc.ChangePasswordCodeResend(ctx)
	if err != nil {
		t.Fatalf("ChangePasswordCodeResend() error = %v", err)
	}
	if msg != "password change code sent" {
		t.Fatalf("ChangePasswordCodeResend() message = %q, want password change code sent", msg)
	}
}

func TestAuthService_ChangePasswordCodeResend_TransactionsNotConfigured(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, nil)
	ctx := authCtx(uuid.New())

	_, err := svc.ChangePasswordCodeResend(ctx)
	if !errors.Is(err, authservice.ErrTransactionsNotConfigured) {
		t.Fatalf("ChangePasswordCodeResend() error = %v, want ErrTransactionsNotConfigured", err)
	}
}

func TestAuthService_ChangePasswordCodeResend_RateLimiterNotConfigured(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newServiceWithoutRateLimiter(deps, transactionWithMocks(deps))
	ctx := authCtx(uuid.New())

	_, err := svc.ChangePasswordCodeResend(ctx)
	if !errors.Is(err, authservice.ErrRateLimiterNotConfigured) {
		t.Fatalf("ChangePasswordCodeResend() error = %v, want ErrRateLimiterNotConfigured", err)
	}
}

func TestAuthService_ChangePasswordCodeResend_IdentityNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(nil, pgx.ErrNoRows)

	_, err := svc.ChangePasswordCodeResend(ctx)
	if err == nil {
		t.Fatal("ChangePasswordCodeResend() expected error, got nil")
	}
}

func TestAuthService_ChangePasswordCodeResend_IdentityNotActive(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)
	identity := activeIdentity(identityID, "user@example.com")
	identity.Status = "blocked"

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(identity, nil)

	_, err := svc.ChangePasswordCodeResend(ctx)
	if !errors.Is(err, authservice.ErrIdentityNotActive) {
		t.Fatalf("ChangePasswordCodeResend() error = %v, want ErrIdentityNotActive", err)
	}
}

func TestAuthService_ChangePasswordCodeResend_EmailNotVerified(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)
	identity := activeIdentity(identityID, "user@example.com")
	identity.EmailVerified = false

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(identity, nil)

	_, err := svc.ChangePasswordCodeResend(ctx)
	if !errors.Is(err, authservice.ErrEmailNotVerified) {
		t.Fatalf("ChangePasswordCodeResend() error = %v, want ErrEmailNotVerified", err)
	}
}

func TestAuthService_ChangePasswordCodeResend_NoActiveCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(activeIdentity(identityID, "user@example.com"), nil)
	deps.verificationCodes.EXPECT().GetActiveByIdentityIDAndPurpose(ctx, identityID, "password_change").Return(nil, pgx.ErrNoRows)

	_, err := svc.ChangePasswordCodeResend(ctx)
	if !errors.Is(err, authservice.ErrPasswordChangeNotFound) {
		t.Fatalf("ChangePasswordCodeResend() error = %v, want ErrPasswordChangeNotFound", err)
	}
}

func TestAuthService_ChangePasswordCodeResend_RateLimited(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)
	rateErr := errors.New("rate limited")

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(activeIdentity(identityID, "user@example.com"), nil)
	deps.verificationCodes.EXPECT().GetActiveByIdentityIDAndPurpose(ctx, identityID, "password_change").Return(&records.VerificationCode{
		ID:         uuid.New(),
		IdentityID: &identityID,
		Purpose:    "password_change",
	}, nil)
	deps.rateLimiter.EXPECT().Allow(ctx, "rate:verification_code:password_change:"+identityID.String(), time.Minute, 15*time.Minute, int64(5)).Return(rateErr)

	_, err := svc.ChangePasswordCodeResend(ctx)
	if !errors.Is(err, rateErr) {
		t.Fatalf("ChangePasswordCodeResend() error = %v, want rateErr", err)
	}
}

func TestAuthService_ChangePasswordCodeResend_TxError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)
	txErr := errors.New("tx error")

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(activeIdentity(identityID, "user@example.com"), nil)
	deps.verificationCodes.EXPECT().GetActiveByIdentityIDAndPurpose(ctx, identityID, "password_change").Return(&records.VerificationCode{
		ID:         uuid.New(),
		IdentityID: &identityID,
		Purpose:    "password_change",
	}, nil)
	deps.rateLimiter.EXPECT().Allow(ctx, "rate:verification_code:password_change:"+identityID.String(), time.Minute, 15*time.Minute, int64(5)).Return(nil)
	deps.verificationCodes.EXPECT().Create(ctx, gomock.Any()).Return(txErr)

	_, err := svc.ChangePasswordCodeResend(ctx)
	if !errors.Is(err, txErr) {
		t.Fatalf("ChangePasswordCodeResend() error = %v, want txErr", err)
	}
}

func TestAuthService_ChangePasswordCodeResend_PublishError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)
	publishErr := errors.New("publish error")

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(activeIdentity(identityID, "user@example.com"), nil)
	deps.verificationCodes.EXPECT().GetActiveByIdentityIDAndPurpose(ctx, identityID, "password_change").Return(&records.VerificationCode{
		ID:         uuid.New(),
		IdentityID: &identityID,
		Purpose:    "password_change",
	}, nil)
	deps.rateLimiter.EXPECT().Allow(ctx, "rate:verification_code:password_change:"+identityID.String(), time.Minute, 15*time.Minute, int64(5)).Return(nil)
	deps.verificationCodes.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.authEvents.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.events.EXPECT().PublishPasswordChangeCodeSend(ctx, gomock.Any()).Return(publishErr)

	_, err := svc.ChangePasswordCodeResend(ctx)
	if !errors.Is(err, publishErr) {
		t.Fatalf("ChangePasswordCodeResend() error = %v, want publishErr", err)
	}
}
