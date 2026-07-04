// Package handlers
package handlers

import (
	"context"
	"time"

	authservice "auth/internal/services/auth"
)

type authService interface {
	Login(ctx context.Context, input *authservice.LoginInput) (*authservice.TokenOutput, error)
	Refresh(ctx context.Context, input *authservice.RefreshInput) (*authservice.TokenOutput, error)
	Logout(ctx context.Context) error

	Register(ctx context.Context, input *authservice.RegisterInput) (*authservice.RegisterOutput, error)
	RegisterVerify(ctx context.Context, input *authservice.VerifyCodeInput) (string, error)
	ResendVerificationCode(ctx context.Context, input *authservice.ResendCodeInput) (string, error)

	GetMe(ctx context.Context) (*authservice.Identity, error)
	DeleteAccount(ctx context.Context, input *authservice.DeleteAccountInput) (string, error)
	DeleteAccountVerify(ctx context.Context, input *authservice.DeleteAccountVerifyInput) error
	DeleteAccountCodeResend(ctx context.Context) (string, error)

	ChangeEmail(ctx context.Context, input *authservice.ChangeEmailInput) (string, error)
	ChangeEmailVerify(ctx context.Context, input *authservice.ChangeEmailVerifyInput) (*authservice.ChangeEmailVerifyOutput, error)
	ChangeEmailConfirm(ctx context.Context, input *authservice.ChangeEmailConfirmInput) (string, error)
	ChangeEmailComplete(ctx context.Context, input *authservice.ChangeEmailCompleteInput) (string, error)
	ChangeEmailCodeResend(ctx context.Context, input *authservice.ChangeEmailResendInput) (string, error)

	ChangePassword(ctx context.Context, input *authservice.ChangePasswordInput) (string, error)
	ChangePasswordVerify(ctx context.Context, input *authservice.ChangePasswordVerifyInput) (*authservice.ChangePasswordVerifyOutput, error)
	CompletePasswordChange(ctx context.Context, input *authservice.CompletePasswordChangeInput) (string, error)
	ChangePasswordCodeResend(ctx context.Context) (string, error)

	ForgotPassword(ctx context.Context, input *authservice.ForgotPasswordInput) (string, error)
	ForgotPasswordVerify(ctx context.Context, input *authservice.ForgotPasswordVerifyInput) (*authservice.ResetTokenOutput, error)
	ForgotPasswordCodeResend(ctx context.Context, input *authservice.ResendCodeInput) (string, error)
	ResetPassword(ctx context.Context, input *authservice.ResetPasswordInput) (string, error)

	DeviceStart(ctx context.Context) (*authservice.DeviceStartOutput, error)
	DeviceConfirm(ctx context.Context, input *authservice.DeviceConfirmInput) (*authservice.DeviceConfirmOutput, error)
	DeviceToken(ctx context.Context, input *authservice.DeviceTokenInput) (*authservice.DeviceTokenOutput, error)
	CliRefresh(ctx context.Context, input *authservice.CliRefreshInput) (*authservice.CliRefreshOutput, error)
}

type AuthHandlers struct {
	as                authService
	refreshSessionTTL time.Duration
}

func NewAuthHandlers(as authService, refreshSessionTTL time.Duration) *AuthHandlers {
	return &AuthHandlers{
		as:                as,
		refreshSessionTTL: refreshSessionTTL,
	}
}
