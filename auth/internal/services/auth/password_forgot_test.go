package authservice_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/mock/gomock"

	"auth/internal/models/records"
	authservice "auth/internal/services/auth"
	"auth/pkg/utils"
)

func TestAuthService_ForgotPassword_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()
	email := "user@example.com"
	identityID := uuid.New()

	deps.identities.EXPECT().GetByEmail(ctx, email).Return(activeIdentity(identityID, email), nil)
	deps.rateLimiter.EXPECT().Allow(ctx, "rate:verification_code:password_forgot:"+email, time.Minute, 15*time.Minute, int64(5)).Return(nil)
	deps.verificationCodes.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.authEvents.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.events.EXPECT().PublishPasswordResetCodeSend(ctx, gomock.Any()).Return(nil)

	msg, err := svc.ForgotPassword(ctx, &authservice.ForgotPasswordInput{Email: email})
	if err != nil {
		t.Fatalf("ForgotPassword() error = %v", err)
	}
	if msg != "password code sent" {
		t.Fatalf("ForgotPassword() message = %q, want password code sent", msg)
	}
}

func TestAuthService_ForgotPassword_TransactionsNotConfigured(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, nil)

	_, err := svc.ForgotPassword(context.Background(), &authservice.ForgotPasswordInput{Email: "user@example.com"})
	if !errors.Is(err, authservice.ErrTransactionsNotConfigured) {
		t.Fatalf("ForgotPassword() error = %v, want ErrTransactionsNotConfigured", err)
	}
}

func TestAuthService_ForgotPassword_RateLimiterNotConfigured(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newServiceWithoutRateLimiter(deps, transactionWithMocks(deps))

	_, err := svc.ForgotPassword(context.Background(), &authservice.ForgotPasswordInput{Email: "user@example.com"})
	if !errors.Is(err, authservice.ErrRateLimiterNotConfigured) {
		t.Fatalf("ForgotPassword() error = %v, want ErrRateLimiterNotConfigured", err)
	}
}

func TestAuthService_ForgotPassword_IdentityNotFoundNeutral(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()

	deps.identities.EXPECT().GetByEmail(ctx, "user@example.com").Return(nil, pgx.ErrNoRows)

	msg, err := svc.ForgotPassword(ctx, &authservice.ForgotPasswordInput{Email: "user@example.com"})
	if err != nil {
		t.Fatalf("ForgotPassword() error = %v", err)
	}
	if msg != "password code sent" {
		t.Fatalf("ForgotPassword() message = %q, want password code sent", msg)
	}
}

func TestAuthService_ForgotPassword_IdentityNotActiveNeutral(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()
	identity := activeIdentity(uuid.New(), "user@example.com")
	identity.Status = "blocked"

	deps.identities.EXPECT().GetByEmail(ctx, "user@example.com").Return(identity, nil)

	msg, err := svc.ForgotPassword(ctx, &authservice.ForgotPasswordInput{Email: "user@example.com"})
	if err != nil {
		t.Fatalf("ForgotPassword() error = %v", err)
	}
	if msg != "password code sent" {
		t.Fatalf("ForgotPassword() message = %q, want password code sent", msg)
	}
}

func TestAuthService_ForgotPassword_EmailNotVerifiedNeutral(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()
	identity := activeIdentity(uuid.New(), "user@example.com")
	identity.EmailVerified = false

	deps.identities.EXPECT().GetByEmail(ctx, "user@example.com").Return(identity, nil)

	msg, err := svc.ForgotPassword(ctx, &authservice.ForgotPasswordInput{Email: "user@example.com"})
	if err != nil {
		t.Fatalf("ForgotPassword() error = %v", err)
	}
	if msg != "password code sent" {
		t.Fatalf("ForgotPassword() message = %q, want password code sent", msg)
	}
}

func TestAuthService_ForgotPassword_RateLimited(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()
	email := "user@example.com"
	rateErr := errors.New("rate limited")

	deps.identities.EXPECT().GetByEmail(ctx, email).Return(activeIdentity(uuid.New(), email), nil)
	deps.rateLimiter.EXPECT().Allow(ctx, "rate:verification_code:password_forgot:"+email, time.Minute, 15*time.Minute, int64(5)).Return(rateErr)

	_, err := svc.ForgotPassword(ctx, &authservice.ForgotPasswordInput{Email: email})
	if !errors.Is(err, rateErr) {
		t.Fatalf("ForgotPassword() error = %v, want rateErr", err)
	}
}

func TestAuthService_ForgotPassword_TxError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()
	email := "user@example.com"
	txErr := errors.New("tx error")

	deps.identities.EXPECT().GetByEmail(ctx, email).Return(activeIdentity(uuid.New(), email), nil)
	deps.rateLimiter.EXPECT().Allow(ctx, "rate:verification_code:password_forgot:"+email, time.Minute, 15*time.Minute, int64(5)).Return(nil)
	deps.verificationCodes.EXPECT().Create(ctx, gomock.Any()).Return(txErr)

	_, err := svc.ForgotPassword(ctx, &authservice.ForgotPasswordInput{Email: email})
	if !errors.Is(err, txErr) {
		t.Fatalf("ForgotPassword() error = %v, want txErr", err)
	}
}

func TestAuthService_ForgotPassword_PublishError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()
	email := "user@example.com"
	publishErr := errors.New("publish error")

	deps.identities.EXPECT().GetByEmail(ctx, email).Return(activeIdentity(uuid.New(), email), nil)
	deps.rateLimiter.EXPECT().Allow(ctx, "rate:verification_code:password_forgot:"+email, time.Minute, 15*time.Minute, int64(5)).Return(nil)
	deps.verificationCodes.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.authEvents.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.events.EXPECT().PublishPasswordResetCodeSend(ctx, gomock.Any()).Return(publishErr)

	_, err := svc.ForgotPassword(ctx, &authservice.ForgotPasswordInput{Email: email})
	if !errors.Is(err, publishErr) {
		t.Fatalf("ForgotPassword() error = %v, want publishErr", err)
	}
}

func TestAuthService_ForgotPasswordVerify_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()
	identityID := uuid.New()
	codeID := uuid.New()
	email := "user@example.com"
	code := "123456"

	deps.verificationCodes.EXPECT().GetActiveByEmailAndPurpose(ctx, email, "password_forgot").Return(&records.VerificationCode{
		ID:            codeID,
		IdentityID:    &identityID,
		Email:         &email,
		Purpose:       "password_forgot",
		CodeHash:      utils.HashSHA256(code),
		AttemptsCount: 0,
		MaxAttempts:   5,
	}, nil)
	deps.identities.EXPECT().GetByID(ctx, identityID).Return(activeIdentity(identityID, email), nil)
	deps.tokens.EXPECT().GenerateRefreshToken().Return("reset-token", "reset-token-hash", nil)
	deps.passwordResetTokens.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.verificationCodes.EXPECT().Consume(ctx, codeID).Return(nil)
	deps.authEvents.EXPECT().Create(ctx, gomock.Any()).Return(nil)

	out, err := svc.ForgotPasswordVerify(ctx, &authservice.ForgotPasswordVerifyInput{Email: email, Code: code})
	if err != nil {
		t.Fatalf("ForgotPasswordVerify() error = %v", err)
	}
	if out.ResetToken != "reset-token" {
		t.Fatalf("ForgotPasswordVerify().ResetToken = %q, want reset-token", out.ResetToken)
	}
	if out.ExpiresIn <= 0 {
		t.Fatalf("ForgotPasswordVerify().ExpiresIn = %d, want positive", out.ExpiresIn)
	}
}

func TestAuthService_ForgotPasswordVerify_TransactionsNotConfigured(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, nil)

	_, err := svc.ForgotPasswordVerify(context.Background(), &authservice.ForgotPasswordVerifyInput{Email: "user@example.com", Code: "123456"})
	if !errors.Is(err, authservice.ErrTransactionsNotConfigured) {
		t.Fatalf("ForgotPasswordVerify() error = %v, want ErrTransactionsNotConfigured", err)
	}
}

func TestAuthService_ForgotPasswordVerify_CodeNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()

	deps.verificationCodes.EXPECT().GetActiveByEmailAndPurpose(ctx, "user@example.com", "password_forgot").Return(nil, pgx.ErrNoRows)

	_, err := svc.ForgotPasswordVerify(ctx, &authservice.ForgotPasswordVerifyInput{Email: "user@example.com", Code: "123456"})
	if !errors.Is(err, authservice.ErrInvalidCode) {
		t.Fatalf("ForgotPasswordVerify() error = %v, want ErrInvalidCode", err)
	}
}

func TestAuthService_ForgotPasswordVerify_TooManyAttempts(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()
	identityID := uuid.New()
	email := "user@example.com"

	deps.verificationCodes.EXPECT().GetActiveByEmailAndPurpose(ctx, email, "password_forgot").Return(&records.VerificationCode{
		ID:            uuid.New(),
		IdentityID:    &identityID,
		Email:         &email,
		Purpose:       "password_forgot",
		CodeHash:      utils.HashSHA256("123456"),
		AttemptsCount: 5,
		MaxAttempts:   5,
	}, nil)

	_, err := svc.ForgotPasswordVerify(ctx, &authservice.ForgotPasswordVerifyInput{Email: email, Code: "123456"})
	if !errors.Is(err, authservice.ErrTooManyAttempts) {
		t.Fatalf("ForgotPasswordVerify() error = %v, want ErrTooManyAttempts", err)
	}
}

func TestAuthService_ForgotPasswordVerify_InvalidCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()
	identityID := uuid.New()
	codeID := uuid.New()
	email := "user@example.com"

	deps.verificationCodes.EXPECT().GetActiveByEmailAndPurpose(ctx, email, "password_forgot").Return(&records.VerificationCode{
		ID:            codeID,
		IdentityID:    &identityID,
		Email:         &email,
		Purpose:       "password_forgot",
		CodeHash:      utils.HashSHA256("123456"),
		AttemptsCount: 0,
		MaxAttempts:   5,
	}, nil)
	deps.verificationCodes.EXPECT().IncrementAttempts(ctx, codeID).Return(nil)

	_, err := svc.ForgotPasswordVerify(ctx, &authservice.ForgotPasswordVerifyInput{Email: email, Code: "000000"})
	if !errors.Is(err, authservice.ErrInvalidCode) {
		t.Fatalf("ForgotPasswordVerify() error = %v, want ErrInvalidCode", err)
	}
}

func TestAuthService_ForgotPasswordVerify_TxError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()
	identityID := uuid.New()
	email := "user@example.com"
	txErr := errors.New("tx error")

	deps.verificationCodes.EXPECT().GetActiveByEmailAndPurpose(ctx, email, "password_forgot").Return(&records.VerificationCode{
		ID:            uuid.New(),
		IdentityID:    &identityID,
		Email:         &email,
		Purpose:       "password_forgot",
		CodeHash:      utils.HashSHA256("123456"),
		AttemptsCount: 0,
		MaxAttempts:   5,
	}, nil)
	deps.identities.EXPECT().GetByID(ctx, identityID).Return(activeIdentity(identityID, email), nil)
	deps.tokens.EXPECT().GenerateRefreshToken().Return("reset-token", "reset-token-hash", nil)
	deps.passwordResetTokens.EXPECT().Create(ctx, gomock.Any()).Return(txErr)

	_, err := svc.ForgotPasswordVerify(ctx, &authservice.ForgotPasswordVerifyInput{Email: email, Code: "123456"})
	if !errors.Is(err, txErr) {
		t.Fatalf("ForgotPasswordVerify() error = %v, want txErr", err)
	}
}

func TestAuthService_ForgotPasswordCodeResend_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()
	email := "user@example.com"
	identityID := uuid.New()

	deps.identities.EXPECT().GetByEmail(ctx, email).Return(activeIdentity(identityID, email), nil)
	deps.rateLimiter.EXPECT().Allow(ctx, "rate:verification_code:password_forgot:"+email, time.Minute, 15*time.Minute, int64(5)).Return(nil)
	deps.verificationCodes.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.authEvents.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.events.EXPECT().PublishPasswordResetCodeSend(ctx, gomock.Any()).Return(nil)

	msg, err := svc.ForgotPasswordCodeResend(ctx, &authservice.ResendCodeInput{Email: email})
	if err != nil {
		t.Fatalf("ForgotPasswordCodeResend() error = %v", err)
	}
	if msg != "password code sent" {
		t.Fatalf("ForgotPasswordCodeResend() message = %q, want password code sent", msg)
	}
}

func TestAuthService_ForgotPasswordCodeResend_TransactionsNotConfigured(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, nil)

	_, err := svc.ForgotPasswordCodeResend(context.Background(), &authservice.ResendCodeInput{Email: "user@example.com"})
	if !errors.Is(err, authservice.ErrTransactionsNotConfigured) {
		t.Fatalf("ForgotPasswordCodeResend() error = %v, want ErrTransactionsNotConfigured", err)
	}
}

func TestAuthService_ForgotPasswordCodeResend_RateLimiterNotConfigured(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newServiceWithoutRateLimiter(deps, transactionWithMocks(deps))

	_, err := svc.ForgotPasswordCodeResend(context.Background(), &authservice.ResendCodeInput{Email: "user@example.com"})
	if !errors.Is(err, authservice.ErrRateLimiterNotConfigured) {
		t.Fatalf("ForgotPasswordCodeResend() error = %v, want ErrRateLimiterNotConfigured", err)
	}
}

func TestAuthService_ForgotPasswordCodeResend_IdentityNotFoundNeutral(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()

	deps.identities.EXPECT().GetByEmail(ctx, "user@example.com").Return(nil, pgx.ErrNoRows)

	msg, err := svc.ForgotPasswordCodeResend(ctx, &authservice.ResendCodeInput{Email: "user@example.com"})
	if err != nil {
		t.Fatalf("ForgotPasswordCodeResend() error = %v", err)
	}
	if msg != "password code sent" {
		t.Fatalf("ForgotPasswordCodeResend() message = %q, want password code sent", msg)
	}
}

func TestAuthService_ForgotPasswordCodeResend_IdentityNotActiveNeutral(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()
	identity := activeIdentity(uuid.New(), "user@example.com")
	identity.Status = "blocked"

	deps.identities.EXPECT().GetByEmail(ctx, "user@example.com").Return(identity, nil)

	msg, err := svc.ForgotPasswordCodeResend(ctx, &authservice.ResendCodeInput{Email: "user@example.com"})
	if err != nil {
		t.Fatalf("ForgotPasswordCodeResend() error = %v", err)
	}
	if msg != "password code sent" {
		t.Fatalf("ForgotPasswordCodeResend() message = %q, want password code sent", msg)
	}
}

func TestAuthService_ForgotPasswordCodeResend_RateLimited(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()
	email := "user@example.com"
	rateErr := errors.New("rate limited")

	deps.identities.EXPECT().GetByEmail(ctx, email).Return(activeIdentity(uuid.New(), email), nil)
	deps.rateLimiter.EXPECT().Allow(ctx, "rate:verification_code:password_forgot:"+email, time.Minute, 15*time.Minute, int64(5)).Return(rateErr)

	_, err := svc.ForgotPasswordCodeResend(ctx, &authservice.ResendCodeInput{Email: email})
	if !errors.Is(err, rateErr) {
		t.Fatalf("ForgotPasswordCodeResend() error = %v, want rateErr", err)
	}
}

func TestAuthService_ForgotPasswordCodeResend_TxError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()
	email := "user@example.com"
	txErr := errors.New("tx error")

	deps.identities.EXPECT().GetByEmail(ctx, email).Return(activeIdentity(uuid.New(), email), nil)
	deps.rateLimiter.EXPECT().Allow(ctx, "rate:verification_code:password_forgot:"+email, time.Minute, 15*time.Minute, int64(5)).Return(nil)
	deps.verificationCodes.EXPECT().Create(ctx, gomock.Any()).Return(txErr)

	_, err := svc.ForgotPasswordCodeResend(ctx, &authservice.ResendCodeInput{Email: email})
	if !errors.Is(err, txErr) {
		t.Fatalf("ForgotPasswordCodeResend() error = %v, want txErr", err)
	}
}

func TestAuthService_ForgotPasswordCodeResend_PublishError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()
	email := "user@example.com"
	publishErr := errors.New("publish error")

	deps.identities.EXPECT().GetByEmail(ctx, email).Return(activeIdentity(uuid.New(), email), nil)
	deps.rateLimiter.EXPECT().Allow(ctx, "rate:verification_code:password_forgot:"+email, time.Minute, 15*time.Minute, int64(5)).Return(nil)
	deps.verificationCodes.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.authEvents.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.events.EXPECT().PublishPasswordResetCodeSend(ctx, gomock.Any()).Return(publishErr)

	_, err := svc.ForgotPasswordCodeResend(ctx, &authservice.ResendCodeInput{Email: email})
	if !errors.Is(err, publishErr) {
		t.Fatalf("ForgotPasswordCodeResend() error = %v, want publishErr", err)
	}
}

func TestAuthService_ResetPassword_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()
	identityID := uuid.New()
	now := time.Now().UTC()

	deps.passwordResetTokens.EXPECT().GetByTokenHash(ctx, utils.HashSHA256("reset-token")).Return(&records.PasswordResetToken{
		ID:         uuid.New(),
		IdentityID: identityID,
		TokenHash:  "reset-token-hash",
		ExpiresAt:  now.Add(time.Hour),
		CreatedAt:  now,
	}, nil)
	deps.identities.EXPECT().GetByID(ctx, identityID).Return(activeIdentity(identityID, "user@example.com"), nil)
	deps.credentials.EXPECT().UpdatePassword(ctx, identityID, gomock.Any()).Return(nil)
	deps.passwordResetTokens.EXPECT().Consume(ctx, gomock.Any()).Return(nil)
	deps.sessions.EXPECT().RevokeAllByIdentityID(ctx, identityID).Return(nil)
	deps.authEvents.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.events.EXPECT().PublishIdentityUpdated(ctx, gomock.Any()).Return(nil)

	msg, err := svc.ResetPassword(ctx, &authservice.ResetPasswordInput{ResetToken: "reset-token", NewPassword: "new-password"})
	if err != nil {
		t.Fatalf("ResetPassword() error = %v", err)
	}
	if msg != "password reset" {
		t.Fatalf("ResetPassword() message = %q, want password reset", msg)
	}
}

func TestAuthService_ResetPassword_TransactionsNotConfigured(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, nil)

	_, err := svc.ResetPassword(context.Background(), &authservice.ResetPasswordInput{ResetToken: "reset-token", NewPassword: "new-password"})
	if !errors.Is(err, authservice.ErrTransactionsNotConfigured) {
		t.Fatalf("ResetPassword() error = %v, want ErrTransactionsNotConfigured", err)
	}
}

func TestAuthService_ResetPassword_TokenNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()

	deps.passwordResetTokens.EXPECT().GetByTokenHash(ctx, utils.HashSHA256("reset-token")).Return(nil, pgx.ErrNoRows)

	_, err := svc.ResetPassword(ctx, &authservice.ResetPasswordInput{ResetToken: "reset-token", NewPassword: "new-password"})
	if !errors.Is(err, authservice.ErrInvalidResetToken) {
		t.Fatalf("ResetPassword() error = %v, want ErrInvalidResetToken", err)
	}
}

func TestAuthService_ResetPassword_TokenConsumed(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()
	now := time.Now().UTC()
	consumedAt := now.Add(-time.Minute)

	deps.passwordResetTokens.EXPECT().GetByTokenHash(ctx, utils.HashSHA256("reset-token")).Return(&records.PasswordResetToken{
		ID:         uuid.New(),
		IdentityID: uuid.New(),
		TokenHash:  "reset-token-hash",
		ExpiresAt:  now.Add(time.Hour),
		ConsumedAt: &consumedAt,
		CreatedAt:  now,
	}, nil)

	_, err := svc.ResetPassword(ctx, &authservice.ResetPasswordInput{ResetToken: "reset-token", NewPassword: "new-password"})
	if !errors.Is(err, authservice.ErrInvalidResetToken) {
		t.Fatalf("ResetPassword() error = %v, want ErrInvalidResetToken", err)
	}
}

func TestAuthService_ResetPassword_TokenExpired(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()
	now := time.Now().UTC()

	deps.passwordResetTokens.EXPECT().GetByTokenHash(ctx, utils.HashSHA256("reset-token")).Return(&records.PasswordResetToken{
		ID:         uuid.New(),
		IdentityID: uuid.New(),
		TokenHash:  "reset-token-hash",
		ExpiresAt:  now.Add(-time.Hour),
		CreatedAt:  now,
	}, nil)

	_, err := svc.ResetPassword(ctx, &authservice.ResetPasswordInput{ResetToken: "reset-token", NewPassword: "new-password"})
	if !errors.Is(err, authservice.ErrResetTokenExpired) {
		t.Fatalf("ResetPassword() error = %v, want ErrResetTokenExpired", err)
	}
}

func TestAuthService_ResetPassword_TxError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()
	identityID := uuid.New()
	now := time.Now().UTC()
	txErr := errors.New("tx error")

	deps.passwordResetTokens.EXPECT().GetByTokenHash(ctx, utils.HashSHA256("reset-token")).Return(&records.PasswordResetToken{
		ID:         uuid.New(),
		IdentityID: identityID,
		TokenHash:  "reset-token-hash",
		ExpiresAt:  now.Add(time.Hour),
		CreatedAt:  now,
	}, nil)
	deps.identities.EXPECT().GetByID(ctx, identityID).Return(activeIdentity(identityID, "user@example.com"), nil)
	deps.credentials.EXPECT().UpdatePassword(ctx, identityID, gomock.Any()).Return(txErr)

	_, err := svc.ResetPassword(ctx, &authservice.ResetPasswordInput{ResetToken: "reset-token", NewPassword: "new-password"})
	if !errors.Is(err, txErr) {
		t.Fatalf("ResetPassword() error = %v, want txErr", err)
	}
}

func TestAuthService_ResetPassword_PublishError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()
	identityID := uuid.New()
	now := time.Now().UTC()
	publishErr := errors.New("publish error")

	deps.passwordResetTokens.EXPECT().GetByTokenHash(ctx, utils.HashSHA256("reset-token")).Return(&records.PasswordResetToken{
		ID:         uuid.New(),
		IdentityID: identityID,
		TokenHash:  "reset-token-hash",
		ExpiresAt:  now.Add(time.Hour),
		CreatedAt:  now,
	}, nil)
	deps.identities.EXPECT().GetByID(ctx, identityID).Return(activeIdentity(identityID, "user@example.com"), nil)
	deps.credentials.EXPECT().UpdatePassword(ctx, identityID, gomock.Any()).Return(nil)
	deps.passwordResetTokens.EXPECT().Consume(ctx, gomock.Any()).Return(nil)
	deps.sessions.EXPECT().RevokeAllByIdentityID(ctx, identityID).Return(nil)
	deps.authEvents.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.events.EXPECT().PublishIdentityUpdated(ctx, gomock.Any()).Return(publishErr)

	_, err := svc.ResetPassword(ctx, &authservice.ResetPasswordInput{ResetToken: "reset-token", NewPassword: "new-password"})
	if !errors.Is(err, publishErr) {
		t.Fatalf("ResetPassword() error = %v, want publishErr", err)
	}
}
