package authservice

import (
	"net"
	"time"
)

// Register

type RegisterInput struct {
	Email    string
	Password string
}

type RegisterOutput struct {
	IdentityID string
	Email      string
	Status     string
}

// Verify / Resend

type VerifyCodeInput struct {
	Email string
	Code  string
}

type ResendCodeInput struct {
	Email string
}

// Login

type LoginInput struct {
	Email     string
	Password  string
	UserAgent string
	IPAddress net.IP
}

type TokenOutput struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int32
}

type LoginOutput = TokenOutput

// Refresh

type RefreshInput struct {
	RefreshToken string
}

type RefreshOutput = TokenOutput

// Me

type Identity struct {
	ID            string
	Email         string
	EmailVerified bool
	Status        string
	CreatedAt     time.Time
}

// Delete Account

type DeleteAccountInput struct {
	Password string
}

type DeleteAccountVerifyInput struct {
	Code string
}

// Email Change

type ChangeEmailInput struct {
	Password string
}

type ChangeEmailVerifyInput struct {
	Code string
}

type ChangeEmailVerifyOutput struct {
	IdentityToken string
	ExpiresIn     int32
}

type ChangeEmailConfirmInput struct {
	NewEmail      string
	IdentityToken string
}

type ChangeEmailCompleteInput struct {
	Code string
}

type ChangeEmailResendInput struct {
	Step string
}

// Password Change

type ChangePasswordInput struct {
	CurrentPassword string
}

type ChangePasswordVerifyInput struct {
	Code string
}

type ChangePasswordVerifyOutput struct {
	ChangeToken string
	ExpiresIn   int32
}

type CompletePasswordChangeInput struct {
	ChangeToken string
	NewPassword string
}

// Forgot Password

type ForgotPasswordInput struct {
	Email string
}

type ForgotPasswordVerifyInput struct {
	Email string
	Code  string
}

type ResetTokenOutput struct {
	ResetToken string
	ExpiresIn  int32
}

// Reset Password

type ResetPasswordInput struct {
	ResetToken  string
	NewPassword string
}

// Device Authorization

type DeviceStartOutput struct {
	DeviceCode      string
	UserCode        string
	VerificationURL string
	ExpiresIn       int32
	Interval        int32
}

type DeviceConfirmInput struct {
	UserCode string
}

type DeviceConfirmOutput struct {
	Status string
}

type DeviceTokenInput struct {
	DeviceCode string
}

type DeviceTokenOutput struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int32
}

type CliRefreshInput struct {
	RefreshToken string
}

type CliRefreshOutput struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int32
}
