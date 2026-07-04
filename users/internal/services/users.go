// Package service
package service

import (
	"context"
	"io"
	"time"
	gitauthservice "users/internal/services/git_auth"

	"github.com/google/uuid"

	"users/internal/models/records"
	"users/pkg/logger"
)

type tokenManager interface {
	GenerateAccessToken(userID, sessionID uuid.UUID) (string, time.Time, error)
	GenerateRefreshToken() (string, string, error)
	ValidateAccessToken(tokenString string) (uuid.UUID, uuid.UUID, string, error)
}

type AvatarStorage interface {
	Save(identityID uuid.UUID, filename string, r io.Reader) (objectKey string, url string, err error)
	Delete(objectKey string) error
}

type avatarStorage = AvatarStorage

type UserProfileRepository interface {
	Create(ctx context.Context, profile *records.UserProfile) error
	GetByID(ctx context.Context, id uuid.UUID) (*records.UserProfile, error)
	GetByEmail(ctx context.Context, email string) (*records.UserProfile, error)
	GetByIdentityID(ctx context.Context, identityID uuid.UUID) (*records.UserProfile, error)
	ExistsByDisplayName(ctx context.Context, displayName string) (bool, error)
	Update(ctx context.Context, profile *records.UserProfile) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type userProfileRepository = UserProfileRepository

type ProcessedEventsRepository interface {
	Create(ctx context.Context, event *records.ProcessedEvent) error
	GetByEventID(ctx context.Context, eventID string) (*records.ProcessedEvent, error)
	GetByAggregateID(ctx context.Context, aggregateID uuid.UUID) ([]records.ProcessedEvent, error)
	Update(ctx context.Context, event *records.ProcessedEvent) error
	Delete(ctx context.Context, eventID string) error
}

type GitAuthService interface {
	RegisterGitUser(ctx context.Context, input *gitauthservice.RegisterGitUserInput) (uuid.UUID, error)
	GetGitMe(ctx context.Context, identityID uuid.UUID) (*gitauthservice.GitMeResponse, error)
}

type UsersService struct {
	log          *logger.Logger
	userProfiles userProfileRepository
	avatar       avatarStorage
	tokens       tokenManager
	gitAuth      GitAuthService
}

func NewUsersService(
	log *logger.Logger,
	userProfiles userProfileRepository,
	avatar avatarStorage,
	tokens tokenManager,
	gitAuth GitAuthService,
) *UsersService {
	return &UsersService{
		log:          log,
		userProfiles: userProfiles,
		avatar:       avatar,
		tokens:       tokens,
		gitAuth:      gitAuth,
	}
}
