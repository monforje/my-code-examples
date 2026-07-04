package gitauthservice

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/google/uuid"

	"users/internal/config"
	"users/internal/models/records"
	clientsdto "users/pkg/http_clients/dto"
	"users/pkg/logger"
)

type mockGitAuthClient struct {
	req  *clientsdto.RegisterGitUserRequest
	resp *clientsdto.RegisterGitUserResponse
	err  error
}

func (m *mockGitAuthClient) RegisterGitUser(ctx context.Context, req *clientsdto.RegisterGitUserRequest) (*clientsdto.RegisterGitUserResponse, error) {
	m.req = req
	return m.resp, m.err
}

type mockGitUserRepository struct {
	gitUser *records.GitUser
	err     error
}

func (m *mockGitUserRepository) Create(ctx context.Context, gitUser *records.GitUser) error {
	m.gitUser = gitUser
	return m.err
}

func (m *mockGitUserRepository) GetByProfileID(ctx context.Context, profileID uuid.UUID) ([]*records.GitUser, error) {
	return nil, nil
}

func newTestLogger() *logger.Logger {
	return logger.New(&config.LoggerConfig{Level: -4, Format: config.FormatText, Output: io.Discard})
}

func TestGitAuthService_RegisterGitUser_Success(t *testing.T) {
	profileID := uuid.New()
	client := &mockGitAuthClient{resp: &clientsdto.RegisterGitUserResponse{
		Username: "alice",
		Token:    "git-token",
		GitURL:   "http://gitea.local",
	}}
	repo := &mockGitUserRepository{}
	svc := NewGitAuthService(newTestLogger(), client, repo, nil)

	gotID, err := svc.RegisterGitUser(context.Background(), &RegisterGitUserInput{
		ProfileID: profileID,
		Username:  "alice",
		Email:     "alice@example.com",
	})
	if err != nil {
		t.Fatalf("RegisterGitUser() error = %v", err)
	}
	if gotID == uuid.Nil {
		t.Fatal("RegisterGitUser() returned nil id")
	}
	if client.req == nil || client.req.Username != "alice" || client.req.Email != "alice@example.com" {
		t.Fatalf("client request = %#v", client.req)
	}
	if repo.gitUser == nil {
		t.Fatal("repo Create was not called")
	}
	if repo.gitUser.ID != gotID {
		t.Fatalf("stored id = %v, want %v", repo.gitUser.ID, gotID)
	}
	if repo.gitUser.ProfileID != profileID {
		t.Fatalf("stored profile id = %v, want %v", repo.gitUser.ProfileID, profileID)
	}
	if repo.gitUser.GitToken != "git-token" || repo.gitUser.GitURL != "http://gitea.local" {
		t.Fatalf("stored git data = token %q url %q", repo.gitUser.GitToken, repo.gitUser.GitURL)
	}
}

func TestGitAuthService_RegisterGitUser_ClientError(t *testing.T) {
	clientErr := errors.New("git auth unavailable")
	client := &mockGitAuthClient{err: clientErr}
	repo := &mockGitUserRepository{}
	svc := NewGitAuthService(newTestLogger(), client, repo, nil)

	_, err := svc.RegisterGitUser(context.Background(), &RegisterGitUserInput{
		ProfileID: uuid.New(),
		Username:  "alice",
		Email:     "alice@example.com",
	})
	if err == nil {
		t.Fatal("RegisterGitUser() error = nil, want error")
	}
	if repo.gitUser != nil {
		t.Fatal("repo Create called after client error")
	}
}

func TestGitAuthService_RegisterGitUser_EmptyResponse(t *testing.T) {
	client := &mockGitAuthClient{resp: &clientsdto.RegisterGitUserResponse{Username: "alice"}}
	repo := &mockGitUserRepository{}
	svc := NewGitAuthService(newTestLogger(), client, repo, nil)

	_, err := svc.RegisterGitUser(context.Background(), &RegisterGitUserInput{
		ProfileID: uuid.New(),
		Username:  "alice",
		Email:     "alice@example.com",
	})
	if err == nil {
		t.Fatal("RegisterGitUser() error = nil, want error")
	}
	if repo.gitUser != nil {
		t.Fatal("repo Create called after empty response")
	}
}

func TestGitAuthService_RegisterGitUser_RepoError(t *testing.T) {
	repoErr := errors.New("insert failed")
	client := &mockGitAuthClient{resp: &clientsdto.RegisterGitUserResponse{
		Username: "alice",
		Token:    "git-token",
		GitURL:   "http://gitea.local",
	}}
	repo := &mockGitUserRepository{err: repoErr}
	svc := NewGitAuthService(newTestLogger(), client, repo, nil)

	_, err := svc.RegisterGitUser(context.Background(), &RegisterGitUserInput{
		ProfileID: uuid.New(),
		Username:  "alice",
		Email:     "alice@example.com",
	})
	if err == nil {
		t.Fatal("RegisterGitUser() error = nil, want error")
	}
	if repo.gitUser == nil {
		t.Fatal("repo Create was not called")
	}
}
