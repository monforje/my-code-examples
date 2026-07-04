// Package authservice
package authservice

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"auth/internal/config"
	"auth/internal/events"
	"auth/internal/models/records"
)

var (
	ErrTransactionsNotConfigured = errors.New("transactions not configured")
	ErrEmailAlreadyExists        = errors.New("email already exists")
	ErrIdentityNotFound          = errors.New("identity not found")
	ErrInvalidCode               = errors.New("invalid code")
	ErrTooManyAttempts           = errors.New("too many attempts")
	ErrEmailAlreadyVerified      = errors.New("email already verified")
	ErrRateLimiterNotConfigured  = errors.New("rate limiter not configured")
	ErrInvalidCredentials        = errors.New("invalid credentials")
	ErrIdentityNotActive         = errors.New("identity is not active")
	ErrEmailNotVerified          = errors.New("email is not verified")
	ErrInvalidRefreshToken       = errors.New("invalid refresh token")
	ErrSessionRevoked            = errors.New("session revoked")
	ErrSessionExpired            = errors.New("session expired")
	ErrCurrentPasswordIncorrect  = errors.New("current password incorrect")
	ErrPasswordChangeNotFound    = errors.New("password change not found")
	ErrInvalidChangeToken        = errors.New("invalid password change token")
	ErrInvalidResetToken         = errors.New("invalid reset token")
	ErrResetTokenExpired         = errors.New("reset token expired")
	ErrInvalidEmailChangeToken   = errors.New("invalid email change token")
	ErrEmailChangeTokenExpired   = errors.New("email change token expired")
	ErrAccountDeleteNotFound     = errors.New("account delete request not found")
	ErrIdentityDeleted           = errors.New("identity is deleted")
	ErrDeviceCodeNotFound        = errors.New("device code not found")
	ErrDeviceCodeExpired         = errors.New("device code expired")
	ErrDeviceCodeAlreadyConfirmed = errors.New("device code already confirmed")
	ErrDeviceCodeNotConfirmed    = errors.New("device code not confirmed yet")
	ErrPollTooFrequent           = errors.New("polling too frequent")
)

const (
	identityStatusActive              = "active"
	identityStatusPendingVerification = "pending_verification"

	verificationPurposeRegister           = "register"
	verificationPurposePasswordChange     = "password_change"
	verificationPurposePasswordForgot     = "password_forgot"
	verificationPurposeEmailChangeCurrent = "email_change_current"
	verificationPurposeEmailChangeNew     = "email_change_new"
	verificationPurposeAccountDelete      = "account_delete"

	passwordCodeSentMessage       = "password code sent"
	passwordChangeCodeSentMessage = "password change code sent"
	passwordChangedMessage        = "password changed"
	passwordResetMessage          = "password reset"

	emailChangeCodeSentMessage = "email change code sent"
	emailChangedMessage        = "email changed"

	emailChangeStatusPending    = "pending"
	emailChangeStatusVerified   = "verified"
	emailChangeStatusConfirming = "confirming"

	accountDeleteCodeSentMessage = "account delete code sent"
	accountDeletedMessage        = "account deleted"

	accountDeleteStatusPending  = "pending"
	accountDeleteStatusVerified = "verified"
)

type tokenManager interface {
	GenerateAccessToken(userID, sessionID uuid.UUID) (string, time.Time, error)
	GenerateRefreshToken() (string, string, error)
	ValidateAccessToken(tokenString string) (uuid.UUID, uuid.UUID, string, error)
}

type IdentityRepository interface {
	Create(ctx context.Context, identity *records.Identity) error
	GetByID(ctx context.Context, id uuid.UUID) (*records.Identity, error)
	GetByEmail(ctx context.Context, email string) (*records.Identity, error)
	Update(ctx context.Context, identity *records.Identity) error
	SetEmailVerified(ctx context.Context, id uuid.UUID) error
	SetStatus(ctx context.Context, id uuid.UUID, status string) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
}

type identityRepository = IdentityRepository

type CredentialRepository interface {
	Create(ctx context.Context, credential *records.Credential) error
	GetByIdentityID(ctx context.Context, identityID uuid.UUID) (*records.Credential, error)
	UpdatePassword(ctx context.Context, identityID uuid.UUID, passwordHash string) error
}

type credentialRepository = CredentialRepository

type SessionRepository interface {
	Create(ctx context.Context, session *records.Session) error
	GetByID(ctx context.Context, id uuid.UUID) (*records.Session, error)
	GetByRefreshTokenHash(ctx context.Context, hash string) (*records.Session, error)
	GetActiveByIdentityID(ctx context.Context, identityID uuid.UUID) ([]*records.Session, error)
	Revoke(ctx context.Context, id uuid.UUID) error
	RevokeAllByIdentityID(ctx context.Context, identityID uuid.UUID) error
	DeleteExpired(ctx context.Context) error
}

type sessionRepository = SessionRepository

type VerificationCodeRepository interface {
	Create(ctx context.Context, code *records.VerificationCode) error
	GetByID(ctx context.Context, id uuid.UUID) (*records.VerificationCode, error)
	GetActiveByEmailAndPurpose(ctx context.Context, email, purpose string) (*records.VerificationCode, error)
	GetActiveByIdentityIDAndPurpose(ctx context.Context, identityID uuid.UUID, purpose string) (*records.VerificationCode, error)
	IncrementAttempts(ctx context.Context, id uuid.UUID) error
	Consume(ctx context.Context, id uuid.UUID) error
	DeleteExpired(ctx context.Context) error
}

type verificationCodeRepository = VerificationCodeRepository

type PasswordResetTokenRepository interface {
	Create(ctx context.Context, token *records.PasswordResetToken) error
	GetByID(ctx context.Context, id uuid.UUID) (*records.PasswordResetToken, error)
	GetByTokenHash(ctx context.Context, hash string) (*records.PasswordResetToken, error)
	Consume(ctx context.Context, id uuid.UUID) error
	DeleteExpired(ctx context.Context) error
}

type passwordResetTokenRepository = PasswordResetTokenRepository

type PasswordChangeTokenRepository interface {
	Create(ctx context.Context, token *records.PasswordChangeToken) error
	GetByTokenHash(ctx context.Context, hash string) (*records.PasswordChangeToken, error)
	Consume(ctx context.Context, id uuid.UUID) error
	DeleteExpired(ctx context.Context) error
}

type passwordChangeTokenRepository = PasswordChangeTokenRepository

type EmailChangeRequestRepository interface {
	Create(ctx context.Context, req *records.EmailChangeRequest) error
	GetByID(ctx context.Context, id uuid.UUID) (*records.EmailChangeRequest, error)
	GetActiveByIdentityIDAndStatus(ctx context.Context, identityID uuid.UUID, status string) (*records.EmailChangeRequest, error)
	GetByTokenHash(ctx context.Context, hash string) (*records.EmailChangeRequest, error)
	UpdateTokenHash(ctx context.Context, id uuid.UUID, tokenHash string) error
	UpdateNewEmailAndStatus(ctx context.Context, id uuid.UUID, newEmail string, status string) error
	SetStatus(ctx context.Context, id uuid.UUID, status string) error
	Consume(ctx context.Context, id uuid.UUID) error
}

type emailChangeRequestRepository = EmailChangeRequestRepository

type AccountDeleteRequestRepository interface {
	Create(ctx context.Context, req *records.AccountDeleteRequest) error
	GetByID(ctx context.Context, id uuid.UUID) (*records.AccountDeleteRequest, error)
	GetActiveByIdentityID(ctx context.Context, identityID uuid.UUID) (*records.AccountDeleteRequest, error)
	SetStatus(ctx context.Context, id uuid.UUID, status string) error
}

type accountDeleteRequestRepository = AccountDeleteRequestRepository

type AuthEventRepository interface {
	Create(ctx context.Context, event *records.AuthEvent) error
	GetByID(ctx context.Context, id uuid.UUID) (*records.AuthEvent, error)
	GetByIdentityID(ctx context.Context, identityID uuid.UUID) ([]*records.AuthEvent, error)
	GetByEventType(ctx context.Context, eventType string) ([]*records.AuthEvent, error)
}

type authEventRepository = AuthEventRepository

type DeviceAuthorizationRepository interface {
	Create(ctx context.Context, dac *records.DeviceAuthorizationCode) error
	GetByDeviceCodeHash(ctx context.Context, deviceCodeHash string) (*records.DeviceAuthorizationCode, error)
	GetByUserCode(ctx context.Context, userCode string) (*records.DeviceAuthorizationCode, error)
	Confirm(ctx context.Context, id uuid.UUID, identityID uuid.UUID) error
	UpdateLastPolledAt(ctx context.Context, id uuid.UUID) error
	DeleteExpired(ctx context.Context) error
}

type deviceAuthorizationRepository = DeviceAuthorizationRepository

type Repositories struct {
	Identities                IdentityRepository
	Credentials               CredentialRepository
	Sessions                  SessionRepository
	VerificationCodes         VerificationCodeRepository
	PasswordResetTokens       PasswordResetTokenRepository
	PasswordChangeTokens      PasswordChangeTokenRepository
	EmailChangeRequests       EmailChangeRequestRepository
	AccountDeleteRequests     AccountDeleteRequestRepository
	AuthEvents                AuthEventRepository
	DeviceAuthorizationCodes  DeviceAuthorizationRepository
}

type TransactionFunc func(ctx context.Context, fn func(Repositories) error) error

type rateLimiter interface {
	Allow(ctx context.Context, key string, cooldown, window time.Duration, maxRequests int64) error
}

type eventProducer interface {
	PublishIdentityCreated(ctx context.Context, payload events.IdentityCreatedPayload) error
	PublishIdentityUpdated(ctx context.Context, payload events.IdentityUpdatedPayload) error
	PublishIdentityDeleted(ctx context.Context, payload events.IdentityDeletedPayload) error
	PublishIdentityLogin(ctx context.Context, payload events.IdentityLoginPayload) error
	PublishIdentityLogout(ctx context.Context, payload events.IdentityLogoutPayload) error
	PublishVerificationCodeSend(ctx context.Context, payload events.VerificationCodeSendPayload) error
	PublishPasswordResetCodeSend(ctx context.Context, payload events.VerificationCodeSendPayload) error
	PublishPasswordChangeCodeSend(ctx context.Context, payload events.VerificationCodeSendPayload) error
	PublishEmailChangeCodeSend(ctx context.Context, payload events.VerificationCodeSendPayload) error
	PublishAccountDeleteCodeSend(ctx context.Context, payload events.VerificationCodeSendPayload) error
}

type AuthService struct {
	identities                identityRepository
	credentials               credentialRepository
	sessions                  sessionRepository
	verificationCodes         verificationCodeRepository
	passwordResetTokens       passwordResetTokenRepository
	passwordChangeTokens      passwordChangeTokenRepository
	emailChangeRequests       emailChangeRequestRepository
	accountDeleteRequests     accountDeleteRequestRepository
	authEvents                authEventRepository
	deviceAuthorizationCodes  deviceAuthorizationRepository
	events                    eventProducer
	tokens                    tokenManager
	transactions              TransactionFunc
	rateLimiter               rateLimiter
	features                  config.FeaturesConfig
	verificationURL           string
}

func NewAuthService(
	identities identityRepository,
	credentials credentialRepository,
	sessions sessionRepository,
	verificationCodes verificationCodeRepository,
	passwordResetTokens passwordResetTokenRepository,
	passwordChangeTokens passwordChangeTokenRepository,
	emailChangeRequests emailChangeRequestRepository,
	accountDeleteRequests accountDeleteRequestRepository,
	authEvents authEventRepository,
	deviceAuthorizationCodes deviceAuthorizationRepository,
	events eventProducer,
	tokens tokenManager,
	transactions TransactionFunc,
	rateLimiter rateLimiter,
	features config.FeaturesConfig,
	verificationURL string,
) *AuthService {
	return &AuthService{
		identities:                identities,
		credentials:               credentials,
		sessions:                  sessions,
		verificationCodes:         verificationCodes,
		passwordResetTokens:       passwordResetTokens,
		passwordChangeTokens:      passwordChangeTokens,
		emailChangeRequests:       emailChangeRequests,
		accountDeleteRequests:     accountDeleteRequests,
		authEvents:                authEvents,
		deviceAuthorizationCodes:  deviceAuthorizationCodes,
		events:                    events,
		tokens:                    tokens,
		transactions:              transactions,
		rateLimiter:               rateLimiter,
		features:                  features,
		verificationURL:           verificationURL,
	}
}
