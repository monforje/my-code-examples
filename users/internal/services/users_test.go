package service_test

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"users/internal/authctx"
	"users/internal/config"
	"users/internal/models/records"
	service "users/internal/services"
	gitauthservice "users/internal/services/git_auth"
	"users/internal/services/mocks"
	"users/pkg/logger"
)

type serviceMocks struct {
	userProfiles *mocks.MockUserProfileRepository
	avatar       *mocks.MockAvatarStorage
	tokens       *mocks.MockTokenManager
	gitAuth      *mockGitAuthService
}

type mockGitAuthService struct {
	calls chan *gitauthservice.RegisterGitUserInput
}

func newMockGitAuthService() *mockGitAuthService {
	return &mockGitAuthService{calls: make(chan *gitauthservice.RegisterGitUserInput, 1)}
}

func (m *mockGitAuthService) RegisterGitUser(ctx context.Context, input *gitauthservice.RegisterGitUserInput) (uuid.UUID, error) {
	m.calls <- input
	return uuid.New(), nil
}

func (m *mockGitAuthService) GetGitMe(ctx context.Context, identityID uuid.UUID) (*gitauthservice.GitMeResponse, error) {
	return &gitauthservice.GitMeResponse{
		Username: "test-user",
		GitToken: "test-token",
		GitURL:   "http://gitea.local",
	}, nil
}

func newServiceMocks(ctrl *gomock.Controller) *serviceMocks {
	return &serviceMocks{
		userProfiles: mocks.NewMockUserProfileRepository(ctrl),
		avatar:       mocks.NewMockAvatarStorage(ctrl),
		tokens:       mocks.NewMockTokenManager(ctrl),
		gitAuth:      newMockGitAuthService(),
	}
}

func newService(deps *serviceMocks) *service.UsersService {
	return service.NewUsersService(
		logger.New(&config.LoggerConfig{Level: -4, Format: config.FormatText, Output: io.Discard}),
		deps.userProfiles,
		deps.avatar,
		deps.tokens,
		deps.gitAuth,
	)
}

func authCtx(identityID uuid.UUID) context.Context {
	return authctx.WithAuth(context.Background(), identityID, uuid.New())
}

func testProfile(identityID uuid.UUID) *records.UserProfile {
	now := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)
	return &records.UserProfile{
		ID:              uuid.New(),
		IdentityID:      identityID,
		Email:           "test@example.com",
		DisplayName:     "Test User",
		BIO:             "hello world",
		AvatarURL:       "https://example.com/avatar.png",
		AvatarObjectKey: "avatars/123.png",
		Status:          "active",
		EmailVerified:   false,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

func testProfileNoAvatar(identityID uuid.UUID) *records.UserProfile {
	now := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)
	return &records.UserProfile{
		ID:              uuid.New(),
		IdentityID:      identityID,
		Email:           "test@example.com",
		DisplayName:     "Test User",
		BIO:             "hello world",
		AvatarURL:       "",
		AvatarObjectKey: "",
		Status:          "active",
		EmailVerified:   false,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

func strPtr(s string) *string {
	return &s
}
