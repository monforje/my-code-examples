package authservice_test

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"auth/internal/config"
	"auth/internal/models/records"
	authservice "auth/internal/services/auth"
	"auth/internal/services/auth/mocks"
)

type serviceMocks struct {
	identities                *mocks.MockIdentityRepository
	credentials               *mocks.MockCredentialRepository
	sessions                  *mocks.MockSessionRepository
	verificationCodes         *mocks.MockVerificationCodeRepository
	passwordResetTokens       *mocks.MockPasswordResetTokenRepository
	passwordChangeTokens      *mocks.MockPasswordChangeTokenRepository
	emailChangeRequests       *mocks.MockEmailChangeRequestRepository
	accountDeleteRequests     *mocks.MockAccountDeleteRequestRepository
	authEvents                *mocks.MockAuthEventRepository
	deviceAuthorizationCodes  *mocks.MockDeviceAuthorizationRepository
	events                    *mocks.MockEventProducer
	tokens                    *mocks.MockTokenManager
	rateLimiter               *mocks.MockRateLimiter
}

func newServiceMocks(ctrl *gomock.Controller) *serviceMocks {
	return &serviceMocks{
		identities:                mocks.NewMockIdentityRepository(ctrl),
		credentials:               mocks.NewMockCredentialRepository(ctrl),
		sessions:                  mocks.NewMockSessionRepository(ctrl),
		verificationCodes:         mocks.NewMockVerificationCodeRepository(ctrl),
		passwordResetTokens:       mocks.NewMockPasswordResetTokenRepository(ctrl),
		passwordChangeTokens:      mocks.NewMockPasswordChangeTokenRepository(ctrl),
		emailChangeRequests:       mocks.NewMockEmailChangeRequestRepository(ctrl),
		accountDeleteRequests:     mocks.NewMockAccountDeleteRequestRepository(ctrl),
		authEvents:                mocks.NewMockAuthEventRepository(ctrl),
		deviceAuthorizationCodes:  mocks.NewMockDeviceAuthorizationRepository(ctrl),
		events:                    mocks.NewMockEventProducer(ctrl),
		tokens:                    mocks.NewMockTokenManager(ctrl),
		rateLimiter:               mocks.NewMockRateLimiter(ctrl),
	}
}

func defaultFeatures() config.FeaturesConfig {
	return config.FeaturesConfig{
		AccessTokenTTL:         15 * time.Minute,
		RefreshTokenLen:        32,
		RefreshSessionTTL:      30 * 24 * time.Hour,
		CodeMaxAttempts:        5,
		CodeTTL:                15 * time.Minute,
		CodeResendCooldown:     time.Minute,
		CodeResendWindow:       15 * time.Minute,
		CodeResendMaxRequests:  5,
		PasswordChangeTokenTTL: 15 * time.Minute,
		PasswordResetTokenTTL:  15 * time.Minute,
		EmailChangeTokenTTL:    15 * time.Minute,
	}
}

func newService(deps *serviceMocks, transactions authservice.TransactionFunc) *authservice.AuthService {
	return authservice.NewAuthService(
		deps.identities,
		deps.credentials,
		deps.sessions,
		deps.verificationCodes,
		deps.passwordResetTokens,
		deps.passwordChangeTokens,
		deps.emailChangeRequests,
		deps.accountDeleteRequests,
		deps.authEvents,
		deps.deviceAuthorizationCodes,
		deps.events,
		deps.tokens,
		transactions,
		deps.rateLimiter,
		defaultFeatures(),
		"https://codurity.dev/cli/login",
	)
}

func newServiceWithoutRateLimiter(deps *serviceMocks, transactions authservice.TransactionFunc) *authservice.AuthService {
	return authservice.NewAuthService(
		deps.identities,
		deps.credentials,
		deps.sessions,
		deps.verificationCodes,
		deps.passwordResetTokens,
		deps.passwordChangeTokens,
		deps.emailChangeRequests,
		deps.accountDeleteRequests,
		deps.authEvents,
		deps.deviceAuthorizationCodes,
		deps.events,
		deps.tokens,
		transactions,
		nil,
		defaultFeatures(),
		"https://codurity.dev/cli/login",
	)
}

func transactionWithMocks(deps *serviceMocks) authservice.TransactionFunc {
	return func(ctx context.Context, fn func(authservice.Repositories) error) error {
		return fn(authservice.Repositories{
			Identities:                deps.identities,
			Credentials:               deps.credentials,
			Sessions:                  deps.sessions,
			VerificationCodes:         deps.verificationCodes,
			PasswordResetTokens:       deps.passwordResetTokens,
			PasswordChangeTokens:      deps.passwordChangeTokens,
			EmailChangeRequests:       deps.emailChangeRequests,
			AccountDeleteRequests:     deps.accountDeleteRequests,
			AuthEvents:                deps.authEvents,
			DeviceAuthorizationCodes:  deps.deviceAuthorizationCodes,
		})
	}
}

func activeIdentity(id uuid.UUID, email string) *records.Identity {
	return &records.Identity{
		ID:            id,
		Email:         email,
		EmailVerified: true,
		Status:        "active",
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
}
