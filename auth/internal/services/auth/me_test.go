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

// --- GetMe ---

func TestAuthService_GetMe_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(activeIdentity(identityID, "user@example.com"), nil)

	identity, err := svc.GetMe(ctx)
	if err != nil {
		t.Fatalf("GetMe() error = %v", err)
	}
	if identity.ID != identityID.String() {
		t.Fatalf("GetMe() ID = %q, want %q", identity.ID, identityID)
	}
	if identity.Email != "user@example.com" {
		t.Fatalf("GetMe() Email = %q, want user@example.com", identity.Email)
	}
	if !identity.EmailVerified {
		t.Fatal("GetMe() EmailVerified = false, want true")
	}
	if identity.Status != "active" {
		t.Fatalf("GetMe() Status = %q, want active", identity.Status)
	}
}

func TestAuthService_GetMe_AuthContextMissing(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()

	_, err := svc.GetMe(ctx)
	if err == nil {
		t.Fatal("GetMe() expected error, got nil")
	}
}

func TestAuthService_GetMe_IdentityNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(nil, pgx.ErrNoRows)

	_, err := svc.GetMe(ctx)
	if err == nil {
		t.Fatal("GetMe() expected error, got nil")
	}
}

func TestAuthService_GetMe_IdentityDeleted(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)
	identity := activeIdentity(identityID, "user@example.com")
	identity.Status = "deleted"

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(identity, nil)

	_, err := svc.GetMe(ctx)
	if !errors.Is(err, authservice.ErrIdentityDeleted) {
		t.Fatalf("GetMe() error = %v, want ErrIdentityDeleted", err)
	}
}

func TestAuthService_GetMe_IdentityBlocked(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)
	identity := activeIdentity(identityID, "user@example.com")
	identity.Status = "blocked"

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(identity, nil)

	_, err := svc.GetMe(ctx)
	if !errors.Is(err, authservice.ErrIdentityNotActive) {
		t.Fatalf("GetMe() error = %v, want ErrIdentityNotActive", err)
	}
}

// --- DeleteAccount ---

func TestAuthService_DeleteAccount_Success(t *testing.T) {
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
	deps.accountDeleteRequests.EXPECT().GetActiveByIdentityID(ctx, identityID).Return(nil, pgx.ErrNoRows)
	deps.rateLimiter.EXPECT().Allow(ctx, "rate:verification_code:account_delete:"+identityID.String(), time.Minute, 15*time.Minute, int64(5)).Return(nil)
	deps.accountDeleteRequests.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.verificationCodes.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.authEvents.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.events.EXPECT().PublishAccountDeleteCodeSend(ctx, gomock.Any()).Return(nil)

	msg, err := svc.DeleteAccount(ctx, &authservice.DeleteAccountInput{Password: password})
	if err != nil {
		t.Fatalf("DeleteAccount() error = %v", err)
	}
	if msg != "account delete code sent" {
		t.Fatalf("DeleteAccount() message = %q, want account delete code sent", msg)
	}
}

func TestAuthService_DeleteAccount_SuccessWithExistingRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	requestID := uuid.New()
	ctx := authCtx(identityID)
	email := "user@example.com"
	password := "current-password"
	passwordHash, _ := utils.HashPassword(password)

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(activeIdentity(identityID, email), nil)
	deps.credentials.EXPECT().GetByIdentityID(ctx, identityID).Return(&records.Credential{IdentityID: identityID, PasswordHash: passwordHash}, nil)
	deps.accountDeleteRequests.EXPECT().GetActiveByIdentityID(ctx, identityID).Return(&records.AccountDeleteRequest{
		ID:         requestID,
		IdentityID: identityID,
		Status:     "pending",
	}, nil)
	deps.rateLimiter.EXPECT().Allow(ctx, "rate:verification_code:account_delete:"+identityID.String(), time.Minute, 15*time.Minute, int64(5)).Return(nil)
	deps.verificationCodes.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.authEvents.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.events.EXPECT().PublishAccountDeleteCodeSend(ctx, gomock.Any()).Return(nil)

	msg, err := svc.DeleteAccount(ctx, &authservice.DeleteAccountInput{Password: password})
	if err != nil {
		t.Fatalf("DeleteAccount() error = %v", err)
	}
	if msg != "account delete code sent" {
		t.Fatalf("DeleteAccount() message = %q, want account delete code sent", msg)
	}
}

func TestAuthService_DeleteAccount_TransactionsNotConfigured(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, nil)
	ctx := authCtx(uuid.New())

	_, err := svc.DeleteAccount(ctx, &authservice.DeleteAccountInput{Password: "password"})
	if !errors.Is(err, authservice.ErrTransactionsNotConfigured) {
		t.Fatalf("DeleteAccount() error = %v, want ErrTransactionsNotConfigured", err)
	}
}

func TestAuthService_DeleteAccount_RateLimiterNotConfigured(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newServiceWithoutRateLimiter(deps, transactionWithMocks(deps))
	ctx := authCtx(uuid.New())

	_, err := svc.DeleteAccount(ctx, &authservice.DeleteAccountInput{Password: "password"})
	if !errors.Is(err, authservice.ErrRateLimiterNotConfigured) {
		t.Fatalf("DeleteAccount() error = %v, want ErrRateLimiterNotConfigured", err)
	}
}

func TestAuthService_DeleteAccount_AuthContextMissing(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()

	_, err := svc.DeleteAccount(ctx, &authservice.DeleteAccountInput{Password: "password"})
	if err == nil {
		t.Fatal("DeleteAccount() expected error, got nil")
	}
}

func TestAuthService_DeleteAccount_IdentityNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(nil, pgx.ErrNoRows)

	_, err := svc.DeleteAccount(ctx, &authservice.DeleteAccountInput{Password: "password"})
	if err == nil {
		t.Fatal("DeleteAccount() expected error, got nil")
	}
}

func TestAuthService_DeleteAccount_IdentityNotActive(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)
	identity := activeIdentity(identityID, "user@example.com")
	identity.Status = "blocked"

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(identity, nil)

	_, err := svc.DeleteAccount(ctx, &authservice.DeleteAccountInput{Password: "password"})
	if !errors.Is(err, authservice.ErrIdentityNotActive) {
		t.Fatalf("DeleteAccount() error = %v, want ErrIdentityNotActive", err)
	}
}

func TestAuthService_DeleteAccount_EmailNotVerified(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)
	identity := activeIdentity(identityID, "user@example.com")
	identity.EmailVerified = false

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(identity, nil)

	_, err := svc.DeleteAccount(ctx, &authservice.DeleteAccountInput{Password: "password"})
	if !errors.Is(err, authservice.ErrEmailNotVerified) {
		t.Fatalf("DeleteAccount() error = %v, want ErrEmailNotVerified", err)
	}
}

func TestAuthService_DeleteAccount_PasswordIncorrect(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)
	passwordHash, _ := utils.HashPassword("correct-password")

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(activeIdentity(identityID, "user@example.com"), nil)
	deps.credentials.EXPECT().GetByIdentityID(ctx, identityID).Return(&records.Credential{IdentityID: identityID, PasswordHash: passwordHash}, nil)

	_, err := svc.DeleteAccount(ctx, &authservice.DeleteAccountInput{Password: "wrong-password"})
	if !errors.Is(err, authservice.ErrCurrentPasswordIncorrect) {
		t.Fatalf("DeleteAccount() error = %v, want ErrCurrentPasswordIncorrect", err)
	}
}

func TestAuthService_DeleteAccount_RateLimited(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)
	passwordHash, _ := utils.HashPassword("password")
	rateErr := errors.New("rate limited")

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(activeIdentity(identityID, "user@example.com"), nil)
	deps.credentials.EXPECT().GetByIdentityID(ctx, identityID).Return(&records.Credential{IdentityID: identityID, PasswordHash: passwordHash}, nil)
	deps.accountDeleteRequests.EXPECT().GetActiveByIdentityID(ctx, identityID).Return(nil, pgx.ErrNoRows)
	deps.rateLimiter.EXPECT().Allow(ctx, "rate:verification_code:account_delete:"+identityID.String(), time.Minute, 15*time.Minute, int64(5)).Return(rateErr)

	_, err := svc.DeleteAccount(ctx, &authservice.DeleteAccountInput{Password: "password"})
	if !errors.Is(err, rateErr) {
		t.Fatalf("DeleteAccount() error = %v, want rateErr", err)
	}
}

func TestAuthService_DeleteAccount_TxError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)
	passwordHash, _ := utils.HashPassword("password")
	txErr := errors.New("tx error")

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(activeIdentity(identityID, "user@example.com"), nil)
	deps.credentials.EXPECT().GetByIdentityID(ctx, identityID).Return(&records.Credential{IdentityID: identityID, PasswordHash: passwordHash}, nil)
	deps.accountDeleteRequests.EXPECT().GetActiveByIdentityID(ctx, identityID).Return(nil, pgx.ErrNoRows)
	deps.rateLimiter.EXPECT().Allow(ctx, "rate:verification_code:account_delete:"+identityID.String(), time.Minute, 15*time.Minute, int64(5)).Return(nil)
	deps.accountDeleteRequests.EXPECT().Create(ctx, gomock.Any()).Return(txErr)

	_, err := svc.DeleteAccount(ctx, &authservice.DeleteAccountInput{Password: "password"})
	if !errors.Is(err, txErr) {
		t.Fatalf("DeleteAccount() error = %v, want txErr", err)
	}
}

func TestAuthService_DeleteAccount_PublishError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)
	email := "user@example.com"
	password := "current-password"
	passwordHash, _ := utils.HashPassword(password)
	publishErr := errors.New("publish error")

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(activeIdentity(identityID, email), nil)
	deps.credentials.EXPECT().GetByIdentityID(ctx, identityID).Return(&records.Credential{IdentityID: identityID, PasswordHash: passwordHash}, nil)
	deps.accountDeleteRequests.EXPECT().GetActiveByIdentityID(ctx, identityID).Return(nil, pgx.ErrNoRows)
	deps.rateLimiter.EXPECT().Allow(ctx, "rate:verification_code:account_delete:"+identityID.String(), time.Minute, 15*time.Minute, int64(5)).Return(nil)
	deps.accountDeleteRequests.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.verificationCodes.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.authEvents.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.events.EXPECT().PublishAccountDeleteCodeSend(ctx, gomock.Any()).Return(publishErr)

	_, err := svc.DeleteAccount(ctx, &authservice.DeleteAccountInput{Password: password})
	if !errors.Is(err, publishErr) {
		t.Fatalf("DeleteAccount() error = %v, want publishErr", err)
	}
}

// --- DeleteAccountVerify ---

func TestAuthService_DeleteAccountVerify_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	requestID := uuid.New()
	codeID := uuid.New()
	ctx := authCtx(identityID)
	code := "123456"

	deps.verificationCodes.EXPECT().GetActiveByIdentityIDAndPurpose(ctx, identityID, "account_delete").Return(&records.VerificationCode{
		ID:            codeID,
		IdentityID:    &identityID,
		Purpose:       "account_delete",
		CodeHash:      utils.HashSHA256(code),
		AttemptsCount: 0,
		MaxAttempts:   5,
	}, nil)
	deps.accountDeleteRequests.EXPECT().GetActiveByIdentityID(ctx, identityID).Return(&records.AccountDeleteRequest{
		ID:         requestID,
		IdentityID: identityID,
		Status:     "pending",
	}, nil)
	deps.accountDeleteRequests.EXPECT().SetStatus(ctx, requestID, "verified").Return(nil)
	deps.verificationCodes.EXPECT().Consume(ctx, codeID).Return(nil)
	deps.identities.EXPECT().SoftDelete(ctx, identityID).Return(nil)
	deps.sessions.EXPECT().RevokeAllByIdentityID(ctx, identityID).Return(nil)
	deps.authEvents.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.events.EXPECT().PublishIdentityDeleted(ctx, gomock.Any()).Return(nil)

	err := svc.DeleteAccountVerify(ctx, &authservice.DeleteAccountVerifyInput{Code: code})
	if err != nil {
		t.Fatalf("DeleteAccountVerify() error = %v", err)
	}
}

func TestAuthService_DeleteAccountVerify_TransactionsNotConfigured(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, nil)
	ctx := authCtx(uuid.New())

	err := svc.DeleteAccountVerify(ctx, &authservice.DeleteAccountVerifyInput{Code: "123456"})
	if !errors.Is(err, authservice.ErrTransactionsNotConfigured) {
		t.Fatalf("DeleteAccountVerify() error = %v, want ErrTransactionsNotConfigured", err)
	}
}

func TestAuthService_DeleteAccountVerify_AuthContextMissing(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()

	err := svc.DeleteAccountVerify(ctx, &authservice.DeleteAccountVerifyInput{Code: "123456"})
	if err == nil {
		t.Fatal("DeleteAccountVerify() expected error, got nil")
	}
}

func TestAuthService_DeleteAccountVerify_CodeNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)

	deps.verificationCodes.EXPECT().GetActiveByIdentityIDAndPurpose(ctx, identityID, "account_delete").Return(nil, pgx.ErrNoRows)

	err := svc.DeleteAccountVerify(ctx, &authservice.DeleteAccountVerifyInput{Code: "123456"})
	if !errors.Is(err, authservice.ErrInvalidCode) {
		t.Fatalf("DeleteAccountVerify() error = %v, want ErrInvalidCode", err)
	}
}

func TestAuthService_DeleteAccountVerify_TooManyAttempts(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)

	deps.verificationCodes.EXPECT().GetActiveByIdentityIDAndPurpose(ctx, identityID, "account_delete").Return(&records.VerificationCode{
		ID:            uuid.New(),
		IdentityID:    &identityID,
		Purpose:       "account_delete",
		CodeHash:      utils.HashSHA256("123456"),
		AttemptsCount: 5,
		MaxAttempts:   5,
	}, nil)

	err := svc.DeleteAccountVerify(ctx, &authservice.DeleteAccountVerifyInput{Code: "123456"})
	if !errors.Is(err, authservice.ErrTooManyAttempts) {
		t.Fatalf("DeleteAccountVerify() error = %v, want ErrTooManyAttempts", err)
	}
}

func TestAuthService_DeleteAccountVerify_InvalidCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	codeID := uuid.New()
	ctx := authCtx(identityID)

	deps.verificationCodes.EXPECT().GetActiveByIdentityIDAndPurpose(ctx, identityID, "account_delete").Return(&records.VerificationCode{
		ID:            codeID,
		IdentityID:    &identityID,
		Purpose:       "account_delete",
		CodeHash:      utils.HashSHA256("123456"),
		AttemptsCount: 0,
		MaxAttempts:   5,
	}, nil)
	deps.verificationCodes.EXPECT().IncrementAttempts(ctx, codeID).Return(nil)

	err := svc.DeleteAccountVerify(ctx, &authservice.DeleteAccountVerifyInput{Code: "000000"})
	if !errors.Is(err, authservice.ErrInvalidCode) {
		t.Fatalf("DeleteAccountVerify() error = %v, want ErrInvalidCode", err)
	}
}

func TestAuthService_DeleteAccountVerify_RequestNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	codeID := uuid.New()
	ctx := authCtx(identityID)

	deps.verificationCodes.EXPECT().GetActiveByIdentityIDAndPurpose(ctx, identityID, "account_delete").Return(&records.VerificationCode{
		ID:            codeID,
		IdentityID:    &identityID,
		Purpose:       "account_delete",
		CodeHash:      utils.HashSHA256("123456"),
		AttemptsCount: 0,
		MaxAttempts:   5,
	}, nil)
	deps.accountDeleteRequests.EXPECT().GetActiveByIdentityID(ctx, identityID).Return(nil, pgx.ErrNoRows)

	err := svc.DeleteAccountVerify(ctx, &authservice.DeleteAccountVerifyInput{Code: "123456"})
	if !errors.Is(err, authservice.ErrAccountDeleteNotFound) {
		t.Fatalf("DeleteAccountVerify() error = %v, want ErrAccountDeleteNotFound", err)
	}
}

func TestAuthService_DeleteAccountVerify_TxError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	requestID := uuid.New()
	codeID := uuid.New()
	ctx := authCtx(identityID)
	txErr := errors.New("tx error")

	deps.verificationCodes.EXPECT().GetActiveByIdentityIDAndPurpose(ctx, identityID, "account_delete").Return(&records.VerificationCode{
		ID:            codeID,
		IdentityID:    &identityID,
		Purpose:       "account_delete",
		CodeHash:      utils.HashSHA256("123456"),
		AttemptsCount: 0,
		MaxAttempts:   5,
	}, nil)
	deps.accountDeleteRequests.EXPECT().GetActiveByIdentityID(ctx, identityID).Return(&records.AccountDeleteRequest{
		ID:         requestID,
		IdentityID: identityID,
		Status:     "pending",
	}, nil)
	deps.accountDeleteRequests.EXPECT().SetStatus(ctx, requestID, "verified").Return(txErr)

	err := svc.DeleteAccountVerify(ctx, &authservice.DeleteAccountVerifyInput{Code: "123456"})
	if !errors.Is(err, txErr) {
		t.Fatalf("DeleteAccountVerify() error = %v, want txErr", err)
	}
}

func TestAuthService_DeleteAccountVerify_PublishError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	requestID := uuid.New()
	codeID := uuid.New()
	ctx := authCtx(identityID)
	publishErr := errors.New("publish error")

	deps.verificationCodes.EXPECT().GetActiveByIdentityIDAndPurpose(ctx, identityID, "account_delete").Return(&records.VerificationCode{
		ID:            codeID,
		IdentityID:    &identityID,
		Purpose:       "account_delete",
		CodeHash:      utils.HashSHA256("123456"),
		AttemptsCount: 0,
		MaxAttempts:   5,
	}, nil)
	deps.accountDeleteRequests.EXPECT().GetActiveByIdentityID(ctx, identityID).Return(&records.AccountDeleteRequest{
		ID:         requestID,
		IdentityID: identityID,
		Status:     "pending",
	}, nil)
	deps.accountDeleteRequests.EXPECT().SetStatus(ctx, requestID, "verified").Return(nil)
	deps.verificationCodes.EXPECT().Consume(ctx, codeID).Return(nil)
	deps.identities.EXPECT().SoftDelete(ctx, identityID).Return(nil)
	deps.sessions.EXPECT().RevokeAllByIdentityID(ctx, identityID).Return(nil)
	deps.authEvents.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.events.EXPECT().PublishIdentityDeleted(ctx, gomock.Any()).Return(publishErr)

	err := svc.DeleteAccountVerify(ctx, &authservice.DeleteAccountVerifyInput{Code: "123456"})
	if !errors.Is(err, publishErr) {
		t.Fatalf("DeleteAccountVerify() error = %v, want publishErr", err)
	}
}

// --- DeleteAccountCodeResend ---

func TestAuthService_DeleteAccountCodeResend_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	requestID := uuid.New()
	ctx := authCtx(identityID)
	email := "user@example.com"

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(activeIdentity(identityID, email), nil)
	deps.accountDeleteRequests.EXPECT().GetActiveByIdentityID(ctx, identityID).Return(&records.AccountDeleteRequest{
		ID:         requestID,
		IdentityID: identityID,
		Status:     "pending",
	}, nil)
	deps.rateLimiter.EXPECT().Allow(ctx, "rate:verification_code:account_delete:"+identityID.String(), time.Minute, 15*time.Minute, int64(5)).Return(nil)
	deps.verificationCodes.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.authEvents.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.events.EXPECT().PublishAccountDeleteCodeSend(ctx, gomock.Any()).Return(nil)

	msg, err := svc.DeleteAccountCodeResend(ctx)
	if err != nil {
		t.Fatalf("DeleteAccountCodeResend() error = %v", err)
	}
	if msg != "account delete code sent" {
		t.Fatalf("DeleteAccountCodeResend() message = %q, want account delete code sent", msg)
	}
}

func TestAuthService_DeleteAccountCodeResend_TransactionsNotConfigured(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, nil)
	ctx := authCtx(uuid.New())

	_, err := svc.DeleteAccountCodeResend(ctx)
	if !errors.Is(err, authservice.ErrTransactionsNotConfigured) {
		t.Fatalf("DeleteAccountCodeResend() error = %v, want ErrTransactionsNotConfigured", err)
	}
}

func TestAuthService_DeleteAccountCodeResend_RateLimiterNotConfigured(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newServiceWithoutRateLimiter(deps, transactionWithMocks(deps))
	ctx := authCtx(uuid.New())

	_, err := svc.DeleteAccountCodeResend(ctx)
	if !errors.Is(err, authservice.ErrRateLimiterNotConfigured) {
		t.Fatalf("DeleteAccountCodeResend() error = %v, want ErrRateLimiterNotConfigured", err)
	}
}

func TestAuthService_DeleteAccountCodeResend_AuthContextMissing(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	ctx := context.Background()

	_, err := svc.DeleteAccountCodeResend(ctx)
	if err == nil {
		t.Fatal("DeleteAccountCodeResend() expected error, got nil")
	}
}

func TestAuthService_DeleteAccountCodeResend_IdentityNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(nil, pgx.ErrNoRows)

	_, err := svc.DeleteAccountCodeResend(ctx)
	if err == nil {
		t.Fatal("DeleteAccountCodeResend() expected error, got nil")
	}
}

func TestAuthService_DeleteAccountCodeResend_IdentityNotActive(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)
	identity := activeIdentity(identityID, "user@example.com")
	identity.Status = "blocked"

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(identity, nil)

	_, err := svc.DeleteAccountCodeResend(ctx)
	if !errors.Is(err, authservice.ErrIdentityNotActive) {
		t.Fatalf("DeleteAccountCodeResend() error = %v, want ErrIdentityNotActive", err)
	}
}

func TestAuthService_DeleteAccountCodeResend_EmailNotVerified(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)
	identity := activeIdentity(identityID, "user@example.com")
	identity.EmailVerified = false

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(identity, nil)

	_, err := svc.DeleteAccountCodeResend(ctx)
	if !errors.Is(err, authservice.ErrEmailNotVerified) {
		t.Fatalf("DeleteAccountCodeResend() error = %v, want ErrEmailNotVerified", err)
	}
}

func TestAuthService_DeleteAccountCodeResend_RequestNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	ctx := authCtx(identityID)

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(activeIdentity(identityID, "user@example.com"), nil)
	deps.accountDeleteRequests.EXPECT().GetActiveByIdentityID(ctx, identityID).Return(nil, pgx.ErrNoRows)

	_, err := svc.DeleteAccountCodeResend(ctx)
	if !errors.Is(err, authservice.ErrAccountDeleteNotFound) {
		t.Fatalf("DeleteAccountCodeResend() error = %v, want ErrAccountDeleteNotFound", err)
	}
}

func TestAuthService_DeleteAccountCodeResend_RateLimited(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	requestID := uuid.New()
	ctx := authCtx(identityID)
	rateErr := errors.New("rate limited")

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(activeIdentity(identityID, "user@example.com"), nil)
	deps.accountDeleteRequests.EXPECT().GetActiveByIdentityID(ctx, identityID).Return(&records.AccountDeleteRequest{
		ID:         requestID,
		IdentityID: identityID,
		Status:     "pending",
	}, nil)
	deps.rateLimiter.EXPECT().Allow(ctx, "rate:verification_code:account_delete:"+identityID.String(), time.Minute, 15*time.Minute, int64(5)).Return(rateErr)

	_, err := svc.DeleteAccountCodeResend(ctx)
	if !errors.Is(err, rateErr) {
		t.Fatalf("DeleteAccountCodeResend() error = %v, want rateErr", err)
	}
}

func TestAuthService_DeleteAccountCodeResend_TxError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	requestID := uuid.New()
	ctx := authCtx(identityID)
	txErr := errors.New("tx error")

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(activeIdentity(identityID, "user@example.com"), nil)
	deps.accountDeleteRequests.EXPECT().GetActiveByIdentityID(ctx, identityID).Return(&records.AccountDeleteRequest{
		ID:         requestID,
		IdentityID: identityID,
		Status:     "pending",
	}, nil)
	deps.rateLimiter.EXPECT().Allow(ctx, "rate:verification_code:account_delete:"+identityID.String(), time.Minute, 15*time.Minute, int64(5)).Return(nil)
	deps.verificationCodes.EXPECT().Create(ctx, gomock.Any()).Return(txErr)

	_, err := svc.DeleteAccountCodeResend(ctx)
	if !errors.Is(err, txErr) {
		t.Fatalf("DeleteAccountCodeResend() error = %v, want txErr", err)
	}
}

func TestAuthService_DeleteAccountCodeResend_PublishError(t *testing.T) {
	ctrl := gomock.NewController(t)
	deps := newServiceMocks(ctrl)
	svc := newService(deps, transactionWithMocks(deps))
	identityID := uuid.New()
	requestID := uuid.New()
	ctx := authCtx(identityID)
	publishErr := errors.New("publish error")

	deps.identities.EXPECT().GetByID(ctx, identityID).Return(activeIdentity(identityID, "user@example.com"), nil)
	deps.accountDeleteRequests.EXPECT().GetActiveByIdentityID(ctx, identityID).Return(&records.AccountDeleteRequest{
		ID:         requestID,
		IdentityID: identityID,
		Status:     "pending",
	}, nil)
	deps.rateLimiter.EXPECT().Allow(ctx, "rate:verification_code:account_delete:"+identityID.String(), time.Minute, 15*time.Minute, int64(5)).Return(nil)
	deps.verificationCodes.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.authEvents.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	deps.events.EXPECT().PublishAccountDeleteCodeSend(ctx, gomock.Any()).Return(publishErr)

	_, err := svc.DeleteAccountCodeResend(ctx)
	if !errors.Is(err, publishErr) {
		t.Fatalf("DeleteAccountCodeResend() error = %v, want publishErr", err)
	}
}

// --- helpers ---

func authCtxWithSession(identityID, sessionID uuid.UUID) context.Context {
	return authctx.WithAuth(context.Background(), identityID, sessionID)
}
