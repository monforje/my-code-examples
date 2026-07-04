// Package gitauthservice
package gitauthservice

import (
	"context"

	clientsdto "users/pkg/http_clients/dto"

	"github.com/google/uuid"

	"users/internal/models/records"
	"users/pkg/logger"
)

type gitAuthClient interface {
	RegisterGitUser(ctx context.Context, req *clientsdto.RegisterGitUserRequest) (*clientsdto.RegisterGitUserResponse, error)
}

type gitUserRepository interface {
	Create(ctx context.Context, gitUser *records.GitUser) error
	GetByProfileID(ctx context.Context, profileID uuid.UUID) ([]*records.GitUser, error)
}

type userProfileRepository interface {
	GetByIdentityID(ctx context.Context, identityID uuid.UUID) (*records.UserProfile, error)
}

type GitAuthService struct {
	log          *logger.Logger
	client       gitAuthClient
	gitUsers     gitUserRepository
	userProfiles userProfileRepository
}

func NewGitAuthService(log *logger.Logger, client gitAuthClient, gitUsers gitUserRepository, userProfiles userProfileRepository) *GitAuthService {
	return &GitAuthService{
		log:          log,
		client:       client,
		gitUsers:     gitUsers,
		userProfiles: userProfiles,
	}
}

var _ interface {
	RegisterGitUser(context.Context, *RegisterGitUserInput) (uuid.UUID, error)
	GetGitMe(context.Context, uuid.UUID) (*GitMeResponse, error)
} = (*GitAuthService)(nil)
