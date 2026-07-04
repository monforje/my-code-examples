package postgresrepo_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"users/internal/models/records"
	postgres "users/internal/repository/postgres"
)

func TestGitUserRepo_Create(t *testing.T) {
	repo := newTestDB(t)
	gitUserRepo := postgres.NewGitUserRepo(repo)
	cleanupTable(t, repo, "git_users")
	cleanupTable(t, repo, "user_profiles")

	ProfileID := createTestUserProfile(t, repo)
	now := time.Now()

	gitUser := &records.GitUser{
		ID:        uuid.New(),
		ProfileID: ProfileID,
		GitToken:  "ghp_test_token_123",
		GitURL:    "https://github.com/testuser",
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := gitUserRepo.Create(context.Background(), gitUser)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := gitUserRepo.GetByID(context.Background(), gitUser.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.GitToken != gitUser.GitToken {
		t.Errorf("GitToken = %v, want %v", got.GitToken, gitUser.GitToken)
	}
	if got.GitURL != gitUser.GitURL {
		t.Errorf("GitURL = %v, want %v", got.GitURL, gitUser.GitURL)
	}
	if got.ProfileID != ProfileID {
		t.Errorf("ProfileID = %v, want %v", got.ProfileID, ProfileID)
	}
}

func TestGitUserRepo_GetByID_NotFound(t *testing.T) {
	repo := newTestDB(t)
	gitUserRepo := postgres.NewGitUserRepo(repo)

	_, err := gitUserRepo.GetByID(context.Background(), uuid.New())
	if err != pgx.ErrNoRows {
		t.Fatalf("GetByID() error = %v, want pgx.ErrNoRows", err)
	}
}

func TestGitUserRepo_GetByProfileID(t *testing.T) {
	repo := newTestDB(t)
	gitUserRepo := postgres.NewGitUserRepo(repo)
	cleanupTable(t, repo, "git_users")
	cleanupTable(t, repo, "user_profiles")

	ProfileID := createTestUserProfile(t, repo)
	now := time.Now()

	for i := 0; i < 3; i++ {
		err := gitUserRepo.Create(context.Background(), &records.GitUser{
			ID:        uuid.New(),
			ProfileID: ProfileID,
			GitToken:  "token_" + string(rune('a'+i)),
			GitURL:    "https://github.com/user" + string(rune('a'+i)),
			CreatedAt: now,
			UpdatedAt: now,
		})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	gitUsers, err := gitUserRepo.GetByProfileID(context.Background(), ProfileID)
	if err != nil {
		t.Fatalf("GetByProfileID() error = %v", err)
	}

	if len(gitUsers) != 3 {
		t.Fatalf("GetByProfileID() returned %d git users, want 3", len(gitUsers))
	}

	for _, gu := range gitUsers {
		if gu.ProfileID != ProfileID {
			t.Errorf("ProfileID = %v, want %v", gu.ProfileID, ProfileID)
		}
	}
}

func TestGitUserRepo_GetByProfileIDAndGitURL(t *testing.T) {
	repo := newTestDB(t)
	gitUserRepo := postgres.NewGitUserRepo(repo)
	cleanupTable(t, repo, "git_users")
	cleanupTable(t, repo, "user_profiles")

	ProfileID := createTestUserProfile(t, repo)
	now := time.Now()

	targetURL := "https://github.com/target"
	err := gitUserRepo.Create(context.Background(), &records.GitUser{
		ID:        uuid.New(),
		ProfileID: ProfileID,
		GitToken:  "token_target",
		GitURL:    targetURL,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := gitUserRepo.GetByProfileIDAndGitURL(context.Background(), ProfileID, targetURL)
	if err != nil {
		t.Fatalf("GetByProfileIDAndGitURL() error = %v", err)
	}

	if got.GitURL != targetURL {
		t.Errorf("GitURL = %v, want %v", got.GitURL, targetURL)
	}
}

func TestGitUserRepo_GetByProfileIDAndGitURL_NotFound(t *testing.T) {
	repo := newTestDB(t)
	gitUserRepo := postgres.NewGitUserRepo(repo)

	_, err := gitUserRepo.GetByProfileIDAndGitURL(context.Background(), uuid.New(), "https://nonexistent")
	if err != pgx.ErrNoRows {
		t.Fatalf("GetByProfileIDAndGitURL() error = %v, want pgx.ErrNoRows", err)
	}
}

func TestGitUserRepo_Update(t *testing.T) {
	repo := newTestDB(t)
	gitUserRepo := postgres.NewGitUserRepo(repo)
	cleanupTable(t, repo, "git_users")
	cleanupTable(t, repo, "user_profiles")

	ProfileID := createTestUserProfile(t, repo)
	now := time.Now()

	gitUser := &records.GitUser{
		ID:        uuid.New(),
		ProfileID: ProfileID,
		GitToken:  "old_token",
		GitURL:    "https://github.com/old",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := gitUserRepo.Create(context.Background(), gitUser); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	gitUser.GitToken = "new_token"
	gitUser.GitURL = "https://github.com/new"
	gitUser.UpdatedAt = time.Now()

	if err := gitUserRepo.Update(context.Background(), gitUser); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	got, err := gitUserRepo.GetByID(context.Background(), gitUser.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.GitToken != "new_token" {
		t.Errorf("GitToken = %v, want new_token", got.GitToken)
	}
	if got.GitURL != "https://github.com/new" {
		t.Errorf("GitURL = %v, want https://github.com/new", got.GitURL)
	}
}

func TestGitUserRepo_Update_NotFound(t *testing.T) {
	repo := newTestDB(t)
	gitUserRepo := postgres.NewGitUserRepo(repo)

	err := gitUserRepo.Update(context.Background(), &records.GitUser{
		ID:        uuid.New(),
		GitToken:  "token",
		GitURL:    "url",
		UpdatedAt: time.Now(),
	})
	if err != pgx.ErrNoRows {
		t.Fatalf("Update() error = %v, want pgx.ErrNoRows", err)
	}
}

func TestGitUserRepo_Delete(t *testing.T) {
	repo := newTestDB(t)
	gitUserRepo := postgres.NewGitUserRepo(repo)
	cleanupTable(t, repo, "git_users")
	cleanupTable(t, repo, "user_profiles")

	ProfileID := createTestUserProfile(t, repo)
	now := time.Now()

	gitUser := &records.GitUser{
		ID:        uuid.New(),
		ProfileID: ProfileID,
		GitToken:  "token",
		GitURL:    "https://github.com/delete",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := gitUserRepo.Create(context.Background(), gitUser); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := gitUserRepo.Delete(context.Background(), gitUser.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err := gitUserRepo.GetByID(context.Background(), gitUser.ID)
	if err != pgx.ErrNoRows {
		t.Fatalf("GetByID() error = %v, want pgx.ErrNoRows", err)
	}
}

func TestGitUserRepo_Delete_NotFound(t *testing.T) {
	repo := newTestDB(t)
	gitUserRepo := postgres.NewGitUserRepo(repo)

	err := gitUserRepo.Delete(context.Background(), uuid.New())
	if err != pgx.ErrNoRows {
		t.Fatalf("Delete() error = %v, want pgx.ErrNoRows", err)
	}
}

func TestGitUserRepo_DeleteByProfileID(t *testing.T) {
	repo := newTestDB(t)
	gitUserRepo := postgres.NewGitUserRepo(repo)
	cleanupTable(t, repo, "git_users")
	cleanupTable(t, repo, "user_profiles")

	ProfileID := createTestUserProfile(t, repo)
	now := time.Now()

	for i := 0; i < 3; i++ {
		err := gitUserRepo.Create(context.Background(), &records.GitUser{
			ID:        uuid.New(),
			ProfileID: ProfileID,
			GitToken:  "token",
			GitURL:    "https://github.com/user" + string(rune('a'+i)),
			CreatedAt: now,
			UpdatedAt: now,
		})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	if err := gitUserRepo.DeleteByProfileID(context.Background(), ProfileID); err != nil {
		t.Fatalf("DeleteByProfileID() error = %v", err)
	}

	gitUsers, err := gitUserRepo.GetByProfileID(context.Background(), ProfileID)
	if err != nil {
		t.Fatalf("GetByProfileID() error = %v", err)
	}

	if len(gitUsers) != 0 {
		t.Fatalf("GetByProfileID() returned %d git users after delete, want 0", len(gitUsers))
	}
}

func TestGitUserRepo_CascadeDelete(t *testing.T) {
	repo := newTestDB(t)
	gitUserRepo := postgres.NewGitUserRepo(repo)
	cleanupTable(t, repo, "git_users")
	cleanupTable(t, repo, "user_profiles")

	ProfileID := createTestUserProfile(t, repo)
	now := time.Now()

	gitUser := &records.GitUser{
		ID:        uuid.New(),
		ProfileID: ProfileID,
		GitToken:  "token",
		GitURL:    "https://github.com/cascade",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := gitUserRepo.Create(context.Background(), gitUser); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if _, err := repo.Exec(context.Background(), "delete from user_profiles where id = $1", ProfileID); err != nil {
		t.Fatalf("delete user profile error = %v", err)
	}

	_, err := gitUserRepo.GetByID(context.Background(), gitUser.ID)
	if err != pgx.ErrNoRows {
		t.Fatalf("GetByID() error = %v, want pgx.ErrNoRows", err)
	}
}

func createTestUserProfile(t *testing.T, repo *postgres.Repo) uuid.UUID {
	t.Helper()

	now := time.Now().UTC()
	profile := &records.UserProfile{
		ID:         uuid.New(),
		IdentityID: uuid.New(),
		Email:      "git-user-" + uuid.NewString() + "@example.com",
		Status:     "active",
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := postgres.NewUserProfileRepo(repo).Create(context.Background(), profile); err != nil {
		t.Fatalf("create test user profile: %v", err)
	}

	return profile.ID
}
