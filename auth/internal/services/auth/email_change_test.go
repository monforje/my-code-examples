package authservice_test

import (
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

// --- ChangeEmail (Step 1) ---

func TestAuthService_ChangeEmail_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)
	email := "user@example.com"
	password := "current-password"
	passwordHash, _ := utils.HashPassword(password)

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(activeIdentity(identityID, email), nil)
	deps.credentials.EXPECT().GetByIdentityID(ctx, identityID).Return(&records.Credential{IdentityID: identityID, PasswordHash: passwordHash}, nil)
	deps.rateLimiter.EXPECT().Allow(ctx, "rate:verification_code:email_change_current:"+identityID.String(), time.Minute, 15*time.Minute, int64(5)).Return(nil)
	deps.emailChangeRequests.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.verificationCodes.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.authEvents.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.events.EXPECT().PublishEmailChangeCodeSend(ctx, gomock.Any()).Return(nil)

	msg, err := svc.ChangeEmail(ctx, &authservice.ChangeEmailInput{Password: password})
	if err != nil {
		t.Fatalf("ChangeEmail() error = %v", err)
	}
	if msg != "email change code sent" {
		t.Fatalf("ChangeEmail() message = %q, want email change code sent", msg)
	}
}

func TestAuthService_ChangeEmail_TransactionsNotConfigured(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, nil)
	ctx := authCtx(uuid.New())

	_, err := svc.ChangeEmail(ctx, &authservice.ChangeEmailInput{Password: "password"})
	if !errors.Is(err, authservice.ErrTransactionsNotConfigured) {
		t.Fatalf("ChangeEmail() error = %v, want ErrTransactionsNotConfigured", err)
	}
}

func TestAuthService_ChangeEmail_RateLimiterNotConfigured(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newServiceWithoutRateLimiter(deps, transactionWithMocks(deps))
	ctx := authCtx(uuid.New())

	_, err := svc.ChangeEmail(ctx, &authservice.ChangeEmailInput{Password: "password"})
	if !errors.Is(err, authservice.ErrRateLimiterNotConfigured) {
		t.Fatalf("ChangeEmail() error = %v, want ErrRateLimiterNotConfigured", err)
	}
}

func TestAuthService_ChangeEmail_IdentityNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(nil, pgx.ErrNoRows)

	_, err := svc.ChangeEmail(ctx, &authservice.ChangeEmailInput{Password: "password"})
	if err == nil {
		t.Fatal("ChangeEmail() expected error, got nil")
	}
}

func TestAuthService_ChangeEmail_IdentityNotActive(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)
	identity := activeIdentity(identityID, "user@example.com")
	identity.Status = "blocked"

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(identity, nil)

	_, err := svc.ChangeEmail(ctx, &authservice.ChangeEmailInput{Password: "password"})
	if !errors.Is(err, authservice.ErrIdentityNotActive) {
		t.Fatalf("ChangeEmail() error = %v, want ErrIdentityNotActive", err)
	}
}

func TestAuthService_ChangeEmail_EmailNotVerified(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)
	identity := activeIdentity(identityID, "user@example.com")
	identity.EmailVerified = false

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(identity, nil)

	_, err := svc.ChangeEmail(ctx, &authservice.ChangeEmailInput{Password: "password"})
	if !errors.Is(err, authservice.ErrEmailNotVerified) {
		t.Fatalf("ChangeEmail() error = %v, want ErrEmailNotVerified", err)
	}
}

func TestAuthService_ChangeEmail_PasswordIncorrect(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)
	passwordHash, _ := utils.HashPassword("correct-password")

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(activeIdentity(identityID, "user@example.com"), nil)
	deps.credentials.EXPECT().GetByIdentityID(ctx, identityID).Return(&records.Credential{IdentityID: identityID, PasswordHash: passwordHash}, nil)

	_, err := svc.ChangeEmail(ctx, &authservice.ChangeEmailInput{Password: "wrong-password"})
	if !errors.Is(err, authservice.ErrCurrentPasswordIncorrect) {
		t.Fatalf("ChangeEmail() error = %v, want ErrCurrentPasswordIncorrect", err)
	}
}

func TestAuthService_ChangeEmail_RateLimited(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)
	passwordHash, _ := utils.HashPassword("password")
	rateErr := errors.New("rate limited")

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(activeIdentity(identityID, "user@example.com"), nil)
	deps.credentials.EXPECT().GetByIdentityID(ctx, identityID).Return(&records.Credential{IdentityID: identityID, PasswordHash: passwordHash}, nil)
	deps.rateLimiter.EXPECT().Allow(ctx, "rate:verification_code:email_change_current:"+identityID.String(), time.Minute, 15*time.Minute, int64(5)).Return(rateErr)

	_, err := svc.ChangeEmail(ctx, &authservice.ChangeEmailInput{Password: "password"})
	if !errors.Is(err, rateErr) {
		t.Fatalf("ChangeEmail() error = %v, want rateErr", err)
	}
}

func TestAuthService_ChangeEmail_TxError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)
	passwordHash, _ := utils.HashPassword("password")
	txErr := errors.New("tx error")

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(activeIdentity(identityID, "user@example.com"), nil)
	deps.credentials.EXPECT().GetByIdentityID(ctx, identityID).Return(&records.Credential{IdentityID: identityID, PasswordHash: passwordHash}, nil)
	deps.rateLimiter.EXPECT().Allow(ctx, "rate:verification_code:email_change_current:"+identityID.String(), time.Minute, 15*time.Minute, int64(5)).Return(nil)
	deps.emailChangeRequests.EXPECT().Create(ctx, gomock.Any()).Return(txErr)

	_, err := svc.ChangeEmail(ctx, &authservice.ChangeEmailInput{Password: "password"})
	if !errors.Is(err, txErr) {
		t.Fatalf("ChangeEmail() error = %v, want txErr", err)
	}
}

func TestAuthService_ChangeEmail_PublishError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)
	passwordHash, _ := utils.HashPassword("password")
	publishErr := errors.New("publish error")

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(activeIdentity(identityID, "user@example.com"), nil)
	deps.credentials.EXPECT().GetByIdentityID(ctx, identityID).Return(&records.Credential{IdentityID: identityID, PasswordHash: passwordHash}, nil)
	deps.rateLimiter.EXPECT().Allow(ctx, "rate:verification_code:email_change_current:"+identityID.String(), time.Minute, 15*time.Minute, int64(5)).Return(nil)
	deps.emailChangeRequests.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.verificationCodes.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.authEvents.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.events.EXPECT().PublishEmailChangeCodeSend(ctx, gomock.Any()).Return(publishErr)

	_, err := svc.ChangeEmail(ctx, &authservice.ChangeEmailInput{Password: "password"})
	if !errors.Is(err, publishErr) {
		t.Fatalf("ChangeEmail() error = %v, want publishErr", err)
	}
}

// --- ChangeEmailVerify (Step 2) ---

func TestAuthService_ChangeEmailVerify_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	codeID := uuid.New()
	requestID := uuid.New()
	ctx := authCtx(identityID)
	code := "123456"

	deps.verificationCodes.EXPECT().GetActiveByIdentityIDAndPurpose(ctx, identityID, "email_change_current").Return(&records.VerificationCode{
		ID:            codeID,
		IdentityID:    &identityID,
		Purpose:       "email_change_current",
		CodeHash:      utils.HashSHA256(code),
		AttemptsCount: 0,
		MaxAttempts:   5,
	}, nil)
	deps.emailChangeRequests.EXPECT().GetActiveByIdentityIDAndStatus(ctx, identityID, "pending").Return(&records.EmailChangeRequest{
		ID:         requestID,
		IdentityID: identityID,
		Status:     "pending",
	}, nil)
	deps.tokens.EXPECT().GenerateRefreshToken().Return("identity-token", "identity-token-hash", nil)
	deps.emailChangeRequests.EXPECT().UpdateTokenHash(ctx, requestID, "identity-token-hash").Return(nil)
	deps.emailChangeRequests.EXPECT().SetStatus(ctx, requestID, "verified").Return(nil)
	deps.verificationCodes.EXPECT().Consume(ctx, codeID).Return(nil)
	deps.authEvents.EXPECT().Create(ctx, gomock.Any()).Return(nil)

	out, err := svc.ChangeEmailVerify(ctx, &authservice.ChangeEmailVerifyInput{Code: code})
	if err != nil {
		t.Fatalf("ChangeEmailVerify() error = %v", err)
	}
	if out.IdentityToken != "identity-token" {
		t.Fatalf("ChangeEmailVerify().IdentityToken = %q, want identity-token", out.IdentityToken)
	}
	if out.ExpiresIn <= 0 {
		t.Fatalf("ChangeEmailVerify().ExpiresIn = %d, want positive", out.ExpiresIn)
	}
}

func TestAuthService_ChangeEmailVerify_TransactionsNotConfigured(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, nil)
	ctx := authCtx(uuid.New())

	_, err := svc.ChangeEmailVerify(ctx, &authservice.ChangeEmailVerifyInput{Code: "123456"})
	if !errors.Is(err, authservice.ErrTransactionsNotConfigured) {
		t.Fatalf("ChangeEmailVerify() error = %v, want ErrTransactionsNotConfigured", err)
	}
}

func TestAuthService_ChangeEmailVerify_CodeNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)

	deps.verificationCodes.EXPECT().GetActiveByIdentityIDAndPurpose(ctx, identityID, "email_change_current").Return(nil, pgx.ErrNoRows)

	_, err := svc.ChangeEmailVerify(ctx, &authservice.ChangeEmailVerifyInput{Code: "123456"})
	if !errors.Is(err, authservice.ErrInvalidCode) {
		t.Fatalf("ChangeEmailVerify() error = %v, want ErrInvalidCode", err)
	}
}

func TestAuthService_ChangeEmailVerify_TooManyAttempts(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)

	deps.verificationCodes.EXPECT().GetActiveByIdentityIDAndPurpose(ctx, identityID, "email_change_current").Return(&records.VerificationCode{
		ID:            uuid.New(),
		IdentityID:    &identityID,
		Purpose:       "email_change_current",
		CodeHash:      utils.HashSHA256("123456"),
		AttemptsCount: 5,
		MaxAttempts:   5,
	}, nil)

	_, err := svc.ChangeEmailVerify(ctx, &authservice.ChangeEmailVerifyInput{Code: "123456"})
	if !errors.Is(err, authservice.ErrTooManyAttempts) {
		t.Fatalf("ChangeEmailVerify() error = %v, want ErrTooManyAttempts", err)
	}
}

func TestAuthService_ChangeEmailVerify_InvalidCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	codeID := uuid.New()
	ctx := authCtx(identityID)

	deps.verificationCodes.EXPECT().GetActiveByIdentityIDAndPurpose(ctx, identityID, "email_change_current").Return(&records.VerificationCode{
		ID:            codeID,
		IdentityID:    &identityID,
		Purpose:       "email_change_current",
		CodeHash:      utils.HashSHA256("123456"),
		AttemptsCount: 0,
		MaxAttempts:   5,
	}, nil)
	deps.verificationCodes.EXPECT().IncrementAttempts(ctx, codeID).Return(nil)

	_, err := svc.ChangeEmailVerify(ctx, &authservice.ChangeEmailVerifyInput{Code: "000000"})
	if !errors.Is(err, authservice.ErrInvalidCode) {
		t.Fatalf("ChangeEmailVerify() error = %v, want ErrInvalidCode", err)
	}
}

func TestAuthService_ChangeEmailVerify_TxError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	codeID := uuid.New()
	requestID := uuid.New()
	ctx := authCtx(identityID)
	txErr := errors.New("tx error")

	deps.verificationCodes.EXPECT().GetActiveByIdentityIDAndPurpose(ctx, identityID, "email_change_current").Return(&records.VerificationCode{
		ID:            codeID,
		IdentityID:    &identityID,
		Purpose:       "email_change_current",
		CodeHash:      utils.HashSHA256("123456"),
		AttemptsCount: 0,
		MaxAttempts:   5,
	}, nil)
	deps.emailChangeRequests.EXPECT().GetActiveByIdentityIDAndStatus(ctx, identityID, "pending").Return(&records.EmailChangeRequest{
		ID:         requestID,
		IdentityID: identityID,
		Status:     "pending",
	}, nil)
	deps.tokens.EXPECT().GenerateRefreshToken().Return("identity-token", "identity-token-hash", nil)
	deps.emailChangeRequests.EXPECT().UpdateTokenHash(ctx, requestID, "identity-token-hash").Return(txErr)

	_, err := svc.ChangeEmailVerify(ctx, &authservice.ChangeEmailVerifyInput{Code: "123456"})
	if !errors.Is(err, txErr) {
		t.Fatalf("ChangeEmailVerify() error = %v, want txErr", err)
	}
}

// --- ChangeEmailConfirm (Step 3) ---

func TestAuthService_ChangeEmailConfirm_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	requestID := uuid.New()
	ctx := authCtx(identityID)
	newEmail := "new@example.com"

	deps.emailChangeRequests.EXPECT().GetByTokenHash(ctx, utils.HashSHA256("identity-token")).Return(&records.EmailChangeRequest{
		ID:         requestID,
		IdentityID: identityID,
		Status:     "verified",
	}, nil)
	deps.identities.EXPECT().GetByEmail(ctx, newEmail).Return(nil, pgx.ErrNoRows)
	deps.rateLimiter.EXPECT().Allow(ctx, "rate:verification_code:email_change_new:"+identityID.String(), time.Minute, 15*time.Minute, int64(5)).Return(nil)
	deps.emailChangeRequests.EXPECT().UpdateNewEmailAndStatus(ctx, requestID, newEmail, "confirming").Return(nil)
	deps.verificationCodes.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.authEvents.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.events.EXPECT().PublishEmailChangeCodeSend(ctx, gomock.Any()).Return(nil)

	msg, err := svc.ChangeEmailConfirm(ctx, &authservice.ChangeEmailConfirmInput{NewEmail: newEmail, IdentityToken: "identity-token"})
	if err != nil {
		t.Fatalf("ChangeEmailConfirm() error = %v", err)
	}
	if msg != "email change code sent" {
		t.Fatalf("ChangeEmailConfirm() message = %q, want email change code sent", msg)
	}
}

func TestAuthService_ChangeEmailConfirm_TransactionsNotConfigured(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, nil)
	ctx := authCtx(uuid.New())

	_, err := svc.ChangeEmailConfirm(ctx, &authservice.ChangeEmailConfirmInput{NewEmail: "new@example.com", IdentityToken: "token"})
	if !errors.Is(err, authservice.ErrTransactionsNotConfigured) {
		t.Fatalf("ChangeEmailConfirm() error = %v, want ErrTransactionsNotConfigured", err)
	}
}

func TestAuthService_ChangeEmailConfirm_RateLimiterNotConfigured(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newServiceWithoutRateLimiter(deps, transactionWithMocks(deps))
	ctx := authCtx(uuid.New())

	_, err := svc.ChangeEmailConfirm(ctx, &authservice.ChangeEmailConfirmInput{NewEmail: "new@example.com", IdentityToken: "token"})
	if !errors.Is(err, authservice.ErrRateLimiterNotConfigured) {
		t.Fatalf("ChangeEmailConfirm() error = %v, want ErrRateLimiterNotConfigured", err)
	}
}

func TestAuthService_ChangeEmailConfirm_TokenNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)

	deps.emailChangeRequests.EXPECT().GetByTokenHash(ctx, utils.HashSHA256("identity-token")).Return(nil, pgx.ErrNoRows)

	_, err := svc.ChangeEmailConfirm(ctx, &authservice.ChangeEmailConfirmInput{NewEmail: "new@example.com", IdentityToken: "identity-token"})
	if !errors.Is(err, authservice.ErrInvalidEmailChangeToken) {
		t.Fatalf("ChangeEmailConfirm() error = %v, want ErrInvalidEmailChangeToken", err)
	}
}

func TestAuthService_ChangeEmailConfirm_TokenWrongIdentity(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)

	deps.emailChangeRequests.EXPECT().GetByTokenHash(ctx, utils.HashSHA256("identity-token")).Return(&records.EmailChangeRequest{
		ID:         uuid.New(),
		IdentityID: uuid.New(),
		Status:     "verified",
	}, nil)

	_, err := svc.ChangeEmailConfirm(ctx, &authservice.ChangeEmailConfirmInput{NewEmail: "new@example.com", IdentityToken: "identity-token"})
	if !errors.Is(err, authservice.ErrInvalidEmailChangeToken) {
		t.Fatalf("ChangeEmailConfirm() error = %v, want ErrInvalidEmailChangeToken", err)
	}
}

func TestAuthService_ChangeEmailConfirm_TokenWrongStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)

	deps.emailChangeRequests.EXPECT().GetByTokenHash(ctx, utils.HashSHA256("identity-token")).Return(&records.EmailChangeRequest{
		ID:         uuid.New(),
		IdentityID: identityID,
		Status:     "pending",
	}, nil)

	_, err := svc.ChangeEmailConfirm(ctx, &authservice.ChangeEmailConfirmInput{NewEmail: "new@example.com", IdentityToken: "identity-token"})
	if !errors.Is(err, authservice.ErrInvalidEmailChangeToken) {
		t.Fatalf("ChangeEmailConfirm() error = %v, want ErrInvalidEmailChangeToken", err)
	}
}

func TestAuthService_ChangeEmailConfirm_NewEmailAlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)
	newEmail := "taken@example.com"

	deps.emailChangeRequests.EXPECT().GetByTokenHash(ctx, utils.HashSHA256("identity-token")).Return(&records.EmailChangeRequest{
		ID:         uuid.New(),
		IdentityID: identityID,
		Status:     "verified",
	}, nil)
	deps.identities.EXPECT().GetByEmail(ctx, newEmail).Return(&records.Identity{ID: uuid.New(), Email: newEmail}, nil)

	_, err := svc.ChangeEmailConfirm(ctx, &authservice.ChangeEmailConfirmInput{NewEmail: newEmail, IdentityToken: "identity-token"})
	if !errors.Is(err, authservice.ErrEmailAlreadyExists) {
		t.Fatalf("ChangeEmailConfirm() error = %v, want ErrEmailAlreadyExists", err)
	}
}

func TestAuthService_ChangeEmailConfirm_RateLimited(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)
	newEmail := "new@example.com"
	rateErr := errors.New("rate limited")

	deps.emailChangeRequests.EXPECT().GetByTokenHash(ctx, utils.HashSHA256("identity-token")).Return(&records.EmailChangeRequest{
		ID:         uuid.New(),
		IdentityID: identityID,
		Status:     "verified",
	}, nil)
	deps.identities.EXPECT().GetByEmail(ctx, newEmail).Return(nil, pgx.ErrNoRows)
	deps.rateLimiter.EXPECT().Allow(ctx, "rate:verification_code:email_change_new:"+identityID.String(), time.Minute, 15*time.Minute, int64(5)).Return(rateErr)

	_, err := svc.ChangeEmailConfirm(ctx, &authservice.ChangeEmailConfirmInput{NewEmail: newEmail, IdentityToken: "identity-token"})
	if !errors.Is(err, rateErr) {
		t.Fatalf("ChangeEmailConfirm() error = %v, want rateErr", err)
	}
}

func TestAuthService_ChangeEmailConfirm_TxError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	requestID := uuid.New()
	ctx := authCtx(identityID)
	newEmail := "new@example.com"
	txErr := errors.New("tx error")

	deps.emailChangeRequests.EXPECT().GetByTokenHash(ctx, utils.HashSHA256("identity-token")).Return(&records.EmailChangeRequest{
		ID:         requestID,
		IdentityID: identityID,
		Status:     "verified",
	}, nil)
	deps.identities.EXPECT().GetByEmail(ctx, newEmail).Return(nil, pgx.ErrNoRows)
	deps.rateLimiter.EXPECT().Allow(ctx, "rate:verification_code:email_change_new:"+identityID.String(), time.Minute, 15*time.Minute, int64(5)).Return(nil)
	deps.emailChangeRequests.EXPECT().UpdateNewEmailAndStatus(ctx, requestID, newEmail, "confirming").Return(txErr)

	_, err := svc.ChangeEmailConfirm(ctx, &authservice.ChangeEmailConfirmInput{NewEmail: newEmail, IdentityToken: "identity-token"})
	if !errors.Is(err, txErr) {
		t.Fatalf("ChangeEmailConfirm() error = %v, want txErr", err)
	}
}

func TestAuthService_ChangeEmailConfirm_PublishError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	requestID := uuid.New()
	ctx := authCtx(identityID)
	newEmail := "new@example.com"
	publishErr := errors.New("publish error")

	deps.emailChangeRequests.EXPECT().GetByTokenHash(ctx, utils.HashSHA256("identity-token")).Return(&records.EmailChangeRequest{
		ID:         requestID,
		IdentityID: identityID,
		Status:     "verified",
	}, nil)
	deps.identities.EXPECT().GetByEmail(ctx, newEmail).Return(nil, pgx.ErrNoRows)
	deps.rateLimiter.EXPECT().Allow(ctx, "rate:verification_code:email_change_new:"+identityID.String(), time.Minute, 15*time.Minute, int64(5)).Return(nil)
	deps.emailChangeRequests.EXPECT().UpdateNewEmailAndStatus(ctx, requestID, newEmail, "confirming").Return(nil)
	deps.verificationCodes.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.authEvents.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.events.EXPECT().PublishEmailChangeCodeSend(ctx, gomock.Any()).Return(publishErr)

	_, err := svc.ChangeEmailConfirm(ctx, &authservice.ChangeEmailConfirmInput{NewEmail: newEmail, IdentityToken: "identity-token"})
	if !errors.Is(err, publishErr) {
		t.Fatalf("ChangeEmailConfirm() error = %v, want publishErr", err)
	}
}

// --- ChangeEmailComplete (Step 4) ---

func TestAuthService_ChangeEmailComplete_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	requestID := uuid.New()
	codeID := uuid.New()
	ctx := authCtx(identityID)
	code := "123456"
	now := time.Now().UTC()

	deps.verificationCodes.EXPECT().GetActiveByIdentityIDAndPurpose(ctx, identityID, "email_change_new").Return(&records.VerificationCode{
		ID:            codeID,
		IdentityID:    &identityID,
		Purpose:       "email_change_new",
		CodeHash:      utils.HashSHA256(code),
		AttemptsCount: 0,
		MaxAttempts:   5,
	}, nil)
	deps.emailChangeRequests.EXPECT().GetActiveByIdentityIDAndStatus(ctx, identityID, "confirming").Return(&records.EmailChangeRequest{
		ID:         requestID,
		IdentityID: identityID,
		NewEmail:   "new@example.com",
		Status:     "confirming",
		ExpiresAt:  now.Add(time.Hour),
		CreatedAt:  now,
	}, nil)
	deps.identities.EXPECT().GetByID(ctx, identityID).Return(activeIdentity(identityID, "user@example.com"), nil)
	deps.identities.EXPECT().Update(ctx, gomock.Any()).Return(nil)
	deps.emailChangeRequests.EXPECT().Consume(ctx, requestID).Return(nil)
	deps.verificationCodes.EXPECT().Consume(ctx, codeID).Return(nil)
	deps.sessions.EXPECT().RevokeAllByIdentityID(ctx, identityID).Return(nil)
	deps.authEvents.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.events.EXPECT().PublishIdentityUpdated(ctx, gomock.Any()).Return(nil)

	msg, err := svc.ChangeEmailComplete(ctx, &authservice.ChangeEmailCompleteInput{Code: code})
	if err != nil {
		t.Fatalf("ChangeEmailComplete() error = %v", err)
	}
	if msg != "email changed" {
		t.Fatalf("ChangeEmailComplete() message = %q, want email changed", msg)
	}
}

func TestAuthService_ChangeEmailComplete_TransactionsNotConfigured(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, nil)
	ctx := authCtx(uuid.New())

	_, err := svc.ChangeEmailComplete(ctx, &authservice.ChangeEmailCompleteInput{Code: "123456"})
	if !errors.Is(err, authservice.ErrTransactionsNotConfigured) {
		t.Fatalf("ChangeEmailComplete() error = %v, want ErrTransactionsNotConfigured", err)
	}
}

func TestAuthService_ChangeEmailComplete_CodeNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)

	deps.verificationCodes.EXPECT().GetActiveByIdentityIDAndPurpose(ctx, identityID, "email_change_new").Return(nil, pgx.ErrNoRows)

	_, err := svc.ChangeEmailComplete(ctx, &authservice.ChangeEmailCompleteInput{Code: "123456"})
	if !errors.Is(err, authservice.ErrInvalidCode) {
		t.Fatalf("ChangeEmailComplete() error = %v, want ErrInvalidCode", err)
	}
}

func TestAuthService_ChangeEmailComplete_TooManyAttempts(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)

	deps.verificationCodes.EXPECT().GetActiveByIdentityIDAndPurpose(ctx, identityID, "email_change_new").Return(&records.VerificationCode{
		ID:            uuid.New(),
		IdentityID:    &identityID,
		Purpose:       "email_change_new",
		CodeHash:      utils.HashSHA256("123456"),
		AttemptsCount: 5,
		MaxAttempts:   5,
	}, nil)

	_, err := svc.ChangeEmailComplete(ctx, &authservice.ChangeEmailCompleteInput{Code: "123456"})
	if !errors.Is(err, authservice.ErrTooManyAttempts) {
		t.Fatalf("ChangeEmailComplete() error = %v, want ErrTooManyAttempts", err)
	}
}

func TestAuthService_ChangeEmailComplete_InvalidCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	codeID := uuid.New()
	ctx := authCtx(identityID)

	deps.verificationCodes.EXPECT().GetActiveByIdentityIDAndPurpose(ctx, identityID, "email_change_new").Return(&records.VerificationCode{
		ID:            codeID,
		IdentityID:    &identityID,
		Purpose:       "email_change_new",
		CodeHash:      utils.HashSHA256("123456"),
		AttemptsCount: 0,
		MaxAttempts:   5,
	}, nil)
	deps.verificationCodes.EXPECT().IncrementAttempts(ctx, codeID).Return(nil)

	_, err := svc.ChangeEmailComplete(ctx, &authservice.ChangeEmailCompleteInput{Code: "000000"})
	if !errors.Is(err, authservice.ErrInvalidCode) {
		t.Fatalf("ChangeEmailComplete() error = %v, want ErrInvalidCode", err)
	}
}

func TestAuthService_ChangeEmailComplete_TxError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	requestID := uuid.New()
	codeID := uuid.New()
	ctx := authCtx(identityID)
	now := time.Now().UTC()
	txErr := errors.New("tx error")

	deps.verificationCodes.EXPECT().GetActiveByIdentityIDAndPurpose(ctx, identityID, "email_change_new").Return(&records.VerificationCode{
		ID:            codeID,
		IdentityID:    &identityID,
		Purpose:       "email_change_new",
		CodeHash:      utils.HashSHA256("123456"),
		AttemptsCount: 0,
		MaxAttempts:   5,
	}, nil)
	deps.emailChangeRequests.EXPECT().GetActiveByIdentityIDAndStatus(ctx, identityID, "confirming").Return(&records.EmailChangeRequest{
		ID:         requestID,
		IdentityID: identityID,
		NewEmail:   "new@example.com",
		Status:     "confirming",
		ExpiresAt:  now.Add(time.Hour),
		CreatedAt:  now,
	}, nil)
	deps.identities.EXPECT().GetByID(ctx, identityID).Return(activeIdentity(identityID, "user@example.com"), nil)
	deps.identities.EXPECT().Update(ctx, gomock.Any()).Return(txErr)

	_, err := svc.ChangeEmailComplete(ctx, &authservice.ChangeEmailCompleteInput{Code: "123456"})
	if !errors.Is(err, txErr) {
		t.Fatalf("ChangeEmailComplete() error = %v, want txErr", err)
	}
}

func TestAuthService_ChangeEmailComplete_PublishError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	requestID := uuid.New()
	codeID := uuid.New()
	ctx := authCtx(identityID)
	now := time.Now().UTC()
	publishErr := errors.New("publish error")

	deps.verificationCodes.EXPECT().GetActiveByIdentityIDAndPurpose(ctx, identityID, "email_change_new").Return(&records.VerificationCode{
		ID:            codeID,
		IdentityID:    &identityID,
		Purpose:       "email_change_new",
		CodeHash:      utils.HashSHA256("123456"),
		AttemptsCount: 0,
		MaxAttempts:   5,
	}, nil)
	deps.emailChangeRequests.EXPECT().GetActiveByIdentityIDAndStatus(ctx, identityID, "confirming").Return(&records.EmailChangeRequest{
		ID:         requestID,
		IdentityID: identityID,
		NewEmail:   "new@example.com",
		Status:     "confirming",
		ExpiresAt:  now.Add(time.Hour),
		CreatedAt:  now,
	}, nil)
	deps.identities.EXPECT().GetByID(ctx, identityID).Return(activeIdentity(identityID, "user@example.com"), nil)
	deps.identities.EXPECT().Update(ctx, gomock.Any()).Return(nil)
	deps.emailChangeRequests.EXPECT().Consume(ctx, requestID).Return(nil)
	deps.verificationCodes.EXPECT().Consume(ctx, codeID).Return(nil)
	deps.sessions.EXPECT().RevokeAllByIdentityID(ctx, identityID).Return(nil)
	deps.authEvents.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.events.EXPECT().PublishIdentityUpdated(ctx, gomock.Any()).Return(publishErr)

	_, err := svc.ChangeEmailComplete(ctx, &authservice.ChangeEmailCompleteInput{Code: "123456"})
	if !errors.Is(err, publishErr) {
		t.Fatalf("ChangeEmailComplete() error = %v, want publishErr", err)
	}
}

// --- ChangeEmailCodeResend ---

func TestAuthService_ChangeEmailCodeResend_Current_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	requestID := uuid.New()
	ctx := authCtx(identityID)
	email := "user@example.com"

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(activeIdentity(identityID, email), nil)
	deps.emailChangeRequests.EXPECT().GetActiveByIdentityIDAndStatus(ctx, identityID, "pending").Return(&records.EmailChangeRequest{
		ID:         requestID,
		IdentityID: identityID,
		Status:     "pending",
	}, nil)
	deps.rateLimiter.EXPECT().Allow(ctx, "rate:verification_code:email_change_current:"+identityID.String(), time.Minute, 15*time.Minute, int64(5)).Return(nil)
	deps.verificationCodes.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.authEvents.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.events.EXPECT().PublishEmailChangeCodeSend(ctx, gomock.Any()).Return(nil)

	msg, err := svc.ChangeEmailCodeResend(ctx, &authservice.ChangeEmailResendInput{Step: "current"})
	if err != nil {
		t.Fatalf("ChangeEmailCodeResend() error = %v", err)
	}
	if msg != "email change code sent" {
		t.Fatalf("ChangeEmailCodeResend() message = %q, want email change code sent", msg)
	}
}

func TestAuthService_ChangeEmailCodeResend_New_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	requestID := uuid.New()
	ctx := authCtx(identityID)
	email := "user@example.com"
	newEmail := "new@example.com"

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(activeIdentity(identityID, email), nil)
	deps.emailChangeRequests.EXPECT().GetActiveByIdentityIDAndStatus(ctx, identityID, "confirming").Return(&records.EmailChangeRequest{
		ID:         requestID,
		IdentityID: identityID,
		NewEmail:   newEmail,
		Status:     "confirming",
	}, nil)
	deps.emailChangeRequests.EXPECT().GetActiveByIdentityIDAndStatus(ctx, identityID, "confirming").Return(&records.EmailChangeRequest{
		ID:         requestID,
		IdentityID: identityID,
		NewEmail:   newEmail,
		Status:     "confirming",
	}, nil)
	deps.rateLimiter.EXPECT().Allow(ctx, "rate:verification_code:email_change_new:"+identityID.String(), time.Minute, 15*time.Minute, int64(5)).Return(nil)
	deps.verificationCodes.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.authEvents.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.events.EXPECT().PublishEmailChangeCodeSend(ctx, gomock.Any()).Return(nil)

	msg, err := svc.ChangeEmailCodeResend(ctx, &authservice.ChangeEmailResendInput{Step: "new"})
	if err != nil {
		t.Fatalf("ChangeEmailCodeResend() error = %v", err)
	}
	if msg != "email change code sent" {
		t.Fatalf("ChangeEmailCodeResend() message = %q, want email change code sent", msg)
	}
}

func TestAuthService_ChangeEmailCodeResend_TransactionsNotConfigured(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, nil)
	ctx := authCtx(uuid.New())

	_, err := svc.ChangeEmailCodeResend(ctx, &authservice.ChangeEmailResendInput{Step: "current"})
	if !errors.Is(err, authservice.ErrTransactionsNotConfigured) {
		t.Fatalf("ChangeEmailCodeResend() error = %v, want ErrTransactionsNotConfigured", err)
	}
}

func TestAuthService_ChangeEmailCodeResend_RateLimiterNotConfigured(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newServiceWithoutRateLimiter(deps, transactionWithMocks(deps))
	ctx := authCtx(uuid.New())

	_, err := svc.ChangeEmailCodeResend(ctx, &authservice.ChangeEmailResendInput{Step: "current"})
	if !errors.Is(err, authservice.ErrRateLimiterNotConfigured) {
		t.Fatalf("ChangeEmailCodeResend() error = %v, want ErrRateLimiterNotConfigured", err)
	}
}

func TestAuthService_ChangeEmailCodeResend_IdentityNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(nil, pgx.ErrNoRows)

	_, err := svc.ChangeEmailCodeResend(ctx, &authservice.ChangeEmailResendInput{Step: "current"})
	if err == nil {
		t.Fatal("ChangeEmailCodeResend() expected error, got nil")
	}
}

func TestAuthService_ChangeEmailCodeResend_IdentityNotActive(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)
	identity := activeIdentity(identityID, "user@example.com")
	identity.Status = "blocked"

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(identity, nil)

	_, err := svc.ChangeEmailCodeResend(ctx, &authservice.ChangeEmailResendInput{Step: "current"})
	if !errors.Is(err, authservice.ErrIdentityNotActive) {
		t.Fatalf("ChangeEmailCodeResend() error = %v, want ErrIdentityNotActive", err)
	}
}

func TestAuthService_ChangeEmailCodeResend_EmailNotVerified(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)
	identity := activeIdentity(identityID, "user@example.com")
	identity.EmailVerified = false

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(identity, nil)

	_, err := svc.ChangeEmailCodeResend(ctx, &authservice.ChangeEmailResendInput{Step: "current"})
	if !errors.Is(err, authservice.ErrEmailNotVerified) {
		t.Fatalf("ChangeEmailCodeResend() error = %v, want ErrEmailNotVerified", err)
	}
}

func TestAuthService_ChangeEmailCodeResend_NoActiveRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(activeIdentity(identityID, "user@example.com"), nil)
	deps.emailChangeRequests.EXPECT().GetActiveByIdentityIDAndStatus(ctx, identityID, "pending").Return(nil, pgx.ErrNoRows)

	_, err := svc.ChangeEmailCodeResend(ctx, &authservice.ChangeEmailResendInput{Step: "current"})
	if !errors.Is(err, authservice.ErrInvalidEmailChangeToken) {
		t.Fatalf("ChangeEmailCodeResend() error = %v, want ErrInvalidEmailChangeToken", err)
	}
}

func TestAuthService_ChangeEmailCodeResend_InvalidStep(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(activeIdentity(identityID, "user@example.com"), nil)

	_, err := svc.ChangeEmailCodeResend(ctx, &authservice.ChangeEmailResendInput{Step: "invalid"})
	if !errors.Is(err, authservice.ErrInvalidCode) {
		t.Fatalf("ChangeEmailCodeResend() error = %v, want ErrInvalidCode", err)
	}
}

func TestAuthService_ChangeEmailCodeResend_RateLimited(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	requestID := uuid.New()
	ctx := authCtx(identityID)
	rateErr := errors.New("rate limited")

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(activeIdentity(identityID, "user@example.com"), nil)
	deps.emailChangeRequests.EXPECT().GetActiveByIdentityIDAndStatus(ctx, identityID, "pending").Return(&records.EmailChangeRequest{
		ID:         requestID,
		IdentityID: identityID,
		Status:     "pending",
	}, nil)
	deps.rateLimiter.EXPECT().Allow(ctx, "rate:verification_code:email_change_current:"+identityID.String(), time.Minute, 15*time.Minute, int64(5)).Return(rateErr)

	_, err := svc.ChangeEmailCodeResend(ctx, &authservice.ChangeEmailResendInput{Step: "current"})
	if !errors.Is(err, rateErr) {
		t.Fatalf("ChangeEmailCodeResend() error = %v, want rateErr", err)
	}
}

func TestAuthService_ChangeEmailCodeResend_TxError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	requestID := uuid.New()
	ctx := authCtx(identityID)
	txErr := errors.New("tx error")

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(activeIdentity(identityID, "user@example.com"), nil)
	deps.emailChangeRequests.EXPECT().GetActiveByIdentityIDAndStatus(ctx, identityID, "pending").Return(&records.EmailChangeRequest{
		ID:         requestID,
		IdentityID: identityID,
		Status:     "pending",
	}, nil)
	deps.rateLimiter.EXPECT().Allow(ctx, "rate:verification_code:email_change_current:"+identityID.String(), time.Minute, 15*time.Minute, int64(5)).Return(nil)
	deps.verificationCodes.EXPECT().Create(ctx, gomock.Any()).Return(txErr)

	_, err := svc.ChangeEmailCodeResend(ctx, &authservice.ChangeEmailResendInput{Step: "current"})
	if !errors.Is(err, txErr) {
		t.Fatalf("ChangeEmailCodeResend() error = %v, want txErr", err)
	}
}

func TestAuthService_ChangeEmailCodeResend_PublishError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	requestID := uuid.New()
	ctx := authCtx(identityID)
	publishErr := errors.New("publish error")

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(activeIdentity(identityID, "user@example.com"), nil)
	deps.emailChangeRequests.EXPECT().GetActiveByIdentityIDAndStatus(ctx, identityID, "pending").Return(&records.EmailChangeRequest{
		ID:         requestID,
		IdentityID: identityID,
		Status:     "pending",
	}, nil)
	deps.rateLimiter.EXPECT().Allow(ctx, "rate:verification_code:email_change_current:"+identityID.String(), time.Minute, 15*time.Minute, int64(5)).Return(nil)
	deps.verificationCodes.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.authEvents.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.events.EXPECT().PublishEmailChangeCodeSend(ctx, gomock.Any()).Return(publishErr)

	_, err := svc.ChangeEmailCodeResend(ctx, &authservice.ChangeEmailResendInput{Step: "current"})
	if !errors.Is(err, publishErr) {
		t.Fatalf("ChangeEmailCodeResend() error = %v, want publishErr", err)
	}
}
