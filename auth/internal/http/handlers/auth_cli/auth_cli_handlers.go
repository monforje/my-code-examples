// Package authclihandlers
package authclihandlers

import (
	"context"
	"time"

	authservice "auth/internal/services/auth"
)

type AuthService interface {
	DeviceStart(ctx context.Context) (*authservice.DeviceStartOutput, error)
	DeviceConfirm(ctx context.Context, input *authservice.DeviceConfirmInput) (*authservice.DeviceConfirmOutput, error)
	DeviceToken(ctx context.Context, input *authservice.DeviceTokenInput) (*authservice.DeviceTokenOutput, error)
	CliRefresh(ctx context.Context, input *authservice.CliRefreshInput) (*authservice.CliRefreshOutput, error)
	Refresh(ctx context.Context, input *authservice.RefreshInput) (*authservice.RefreshOutput, error)
}

type AuthCliHandlers struct {
	as                AuthService
	refreshSessionTTL time.Duration
}

func NewAuthCliHandlers(as AuthService, refreshSessionTTL time.Duration) *AuthCliHandlers {
	return &AuthCliHandlers{
		as:                as,
		refreshSessionTTL: refreshSessionTTL,
	}
}
