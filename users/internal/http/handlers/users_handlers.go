// Package handlers
package handlers

import (
	"context"

	"github.com/google/uuid"

	service "users/internal/services"
)

type usersService interface {
	GetProfile(ctx context.Context) (*service.ProfileOutput, error)
	UpdateSettings(ctx context.Context, input *service.UpdateSettingsInput) (*service.ProfileOutput, error)
	UpdateAvatar(ctx context.Context, input *service.UpdateAvatarInput) (*service.UpdateAvatarOutput, error)
	DeleteAvatar(ctx context.Context) (*service.DeleteAvatarOutput, error)
	GetGitMe(ctx context.Context, identityID uuid.UUID) (*service.GitMeOutput, error)
}

type UsersHandlers struct {
	us usersService
}

func NewUsersHandlers(us usersService) *UsersHandlers {
	return &UsersHandlers{
		us: us,
	}
}
