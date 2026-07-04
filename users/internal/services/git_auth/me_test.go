package gitauthservice

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"users/internal/models/records"
	postgresrepo "users/internal/repository/postgres"
)

type mockGetByProfileIDRepo struct {
	gitUsers []*records.GitUser
	err      error
}

func (m *mockGetByProfileIDRepo) Create(ctx context.Context, gitUser *records.GitUser) error {
	return nil
}

func (m *mockGetByProfileIDRepo) GetByProfileID(ctx context.Context, profileID uuid.UUID) ([]*records.GitUser, error) {
	return m.gitUsers, m.err
}

type mockGetByIdentityIDProfileRepo struct {
	profile *records.UserProfile
	err     error
}

func (m *mockGetByIdentityIDProfileRepo) GetByIdentityID(ctx context.Context, identityID uuid.UUID) (*records.UserProfile, error) {
	return m.profile, m.err
}

func TestGitAuthService_GetGitMe_Success(t *testing.T) {
	identityID := uuid.New()
	profileID := uuid.New()

	profile := &records.UserProfile{
		ID:          profileID,
		IdentityID:  identityID,
		DisplayName: "alice",
		Email:       "alice@example.com",
	}
	gitUsers := []*records.GitUser{
		{
			ID:        uuid.New(),
			ProfileID: profileID,
			GitToken:  "git-token-123",
			GitURL:    "http://gitea.local",
		},
	}

	gitUserRepo := &mockGetByProfileIDRepo{gitUsers: gitUsers}
	profileRepo := &mockGetByIdentityIDProfileRepo{profile: profile}
	svc := NewGitAuthService(newTestLogger(), nil, gitUserRepo, profileRepo)

	resp, err := svc.GetGitMe(context.Background(), identityID)
	if err != nil {
		t.Fatalf("GetGitMe() error = %v", err)
	}
	if resp.Username != "alice" {
		t.Errorf("username = %q, want alice", resp.Username)
	}
	if resp.GitToken != "git-token-123" {
		t.Errorf("git_token = %q, want git-token-123", resp.GitToken)
	}
	if resp.GitURL != "http://gitea.local" {
		t.Errorf("git_url = %q, want http://gitea.local", resp.GitURL)
	}
}

func TestGitAuthService_GetGitMe_FallbackToEmail(t *testing.T) {
	identityID := uuid.New()
	profileID := uuid.New()

	profile := &records.UserProfile{
		ID:          profileID,
		IdentityID:  identityID,
		DisplayName: "",
		Email:       "bob@example.com",
	}
	gitUsers := []*records.GitUser{
		{
			ID:        uuid.New(),
			ProfileID: profileID,
			GitToken:  "token",
			GitURL:    "http://gitea.local",
		},
	}

	gitUserRepo := &mockGetByProfileIDRepo{gitUsers: gitUsers}
	profileRepo := &mockGetByIdentityIDProfileRepo{profile: profile}
	svc := NewGitAuthService(newTestLogger(), nil, gitUserRepo, profileRepo)

	resp, err := svc.GetGitMe(context.Background(), identityID)
	if err != nil {
		t.Fatalf("GetGitMe() error = %v", err)
	}
	if resp.Username != "bob@example.com" {
		t.Errorf("username = %q, want bob@example.com", resp.Username)
	}
}

func TestGitAuthService_GetGitMe_ProfileNotFound(t *testing.T) {
	identityID := uuid.New()

	profileRepo := &mockGetByIdentityIDProfileRepo{err: postgresrepo.ErrUserProfileNotFound}
	gitUserRepo := &mockGetByProfileIDRepo{}
	svc := NewGitAuthService(newTestLogger(), nil, gitUserRepo, profileRepo)

	_, err := svc.GetGitMe(context.Background(), identityID)
	if err == nil {
		t.Fatal("GetGitMe() error = nil, want error")
	}
}

func TestGitAuthService_GetGitMe_NoGitUsers(t *testing.T) {
	identityID := uuid.New()
	profileID := uuid.New()

	profile := &records.UserProfile{
		ID:          profileID,
		IdentityID:  identityID,
		DisplayName: "alice",
	}
	gitUserRepo := &mockGetByProfileIDRepo{gitUsers: []*records.GitUser{}}
	profileRepo := &mockGetByIdentityIDProfileRepo{profile: profile}
	svc := NewGitAuthService(newTestLogger(), nil, gitUserRepo, profileRepo)

	_, err := svc.GetGitMe(context.Background(), identityID)
	if err == nil {
		t.Fatal("GetGitMe() error = nil, want error")
	}
	if !errors.Is(err, postgresrepo.ErrGitUserNotFound) {
		t.Fatalf("GetGitMe() error = %v, want %v", err, postgresrepo.ErrGitUserNotFound)
	}
}

func TestGitAuthService_GetGitMe_GitUsersRepoError(t *testing.T) {
	identityID := uuid.New()
	profileID := uuid.New()

	profile := &records.UserProfile{
		ID:          profileID,
		IdentityID:  identityID,
	}
	gitUserRepo := &mockGetByProfileIDRepo{err: errors.New("db down")}
	profileRepo := &mockGetByIdentityIDProfileRepo{profile: profile}
	svc := NewGitAuthService(newTestLogger(), nil, gitUserRepo, profileRepo)

	_, err := svc.GetGitMe(context.Background(), identityID)
	if err == nil {
		t.Fatal("GetGitMe() error = nil, want error")
	}
}

func TestGitAuthService_GetGitMe_MultipleGitUsers(t *testing.T) {
	identityID := uuid.New()
	profileID := uuid.New()

	profile := &records.UserProfile{
		ID:          profileID,
		IdentityID:  identityID,
		DisplayName: "alice",
	}
	gitUsers := []*records.GitUser{
		{
			ID:        uuid.New(),
			ProfileID: profileID,
			GitToken:  "token-2",
			GitURL:    "http://gitea2.local",
		},
		{
			ID:        uuid.New(),
			ProfileID: profileID,
			GitToken:  "token-1",
			GitURL:    "http://gitea1.local",
		},
	}

	gitUserRepo := &mockGetByProfileIDRepo{gitUsers: gitUsers}
	profileRepo := &mockGetByIdentityIDProfileRepo{profile: profile}
	svc := NewGitAuthService(newTestLogger(), nil, gitUserRepo, profileRepo)

	resp, err := svc.GetGitMe(context.Background(), identityID)
	if err != nil {
		t.Fatalf("GetGitMe() error = %v", err)
	}
	// Should return the first (most recent) git user
	if resp.GitToken != "token-2" {
		t.Errorf("git_token = %q, want token-2", resp.GitToken)
	}
	if resp.GitURL != "http://gitea2.local" {
		t.Errorf("git_url = %q, want http://gitea2.local", resp.GitURL)
	}
}
