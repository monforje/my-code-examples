package postgresrepo_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"users/internal/models/records"
	postgresrepo "users/internal/repository/postgres"
)

func TestUserProfileRepo_Create(t *testing.T) {
	repo := newTestDB(t)
	profileRepo := postgresrepo.NewUserProfileRepo(repo)
	cleanupTable(t, repo, "user_profiles")

	now := time.Now().UTC()
	profile := &records.UserProfile{
		ID:              uuid.New(),
		IdentityID:      uuid.New(),
		Email:           "create@example.com",
		DisplayName:     "Test User",
		BIO:             "hello world",
		AvatarURL:       "https://example.com/avatar.png",
		AvatarObjectKey: "avatars/123.png",
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	err := profileRepo.Create(context.Background(), profile)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := profileRepo.GetByID(context.Background(), profile.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.Email != profile.Email {
		t.Errorf("Email = %v, want %v", got.Email, profile.Email)
	}
	if got.DisplayName != profile.DisplayName {
		t.Errorf("DisplayName = %v, want %v", got.DisplayName, profile.DisplayName)
	}
	if got.BIO != profile.BIO {
		t.Errorf("BIO = %v, want %v", got.BIO, profile.BIO)
	}
}

func TestUserProfileRepo_GetByID_NotFound(t *testing.T) {
	repo := newTestDB(t)
	profileRepo := postgresrepo.NewUserProfileRepo(repo)
	cleanupTable(t, repo, "user_profiles")

	_, err := profileRepo.GetByID(context.Background(), uuid.New())
	if !errors.Is(err, postgresrepo.ErrUserProfileNotFound) {
		t.Fatalf("GetByID() error = %v, want %v", err, postgresrepo.ErrUserProfileNotFound)
	}
}

func TestUserProfileRepo_GetByEmail(t *testing.T) {
	repo := newTestDB(t)
	profileRepo := postgresrepo.NewUserProfileRepo(repo)
	cleanupTable(t, repo, "user_profiles")

	now := time.Now().UTC()
	profile := &records.UserProfile{
		ID:        uuid.New(),
		IdentityID: uuid.New(),
		Email:     "find@example.com",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := profileRepo.Create(context.Background(), profile); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := profileRepo.GetByEmail(context.Background(), "find@example.com")
	if err != nil {
		t.Fatalf("GetByEmail() error = %v", err)
	}

	if got.ID != profile.ID {
		t.Errorf("ID = %v, want %v", got.ID, profile.ID)
	}
}

func TestUserProfileRepo_GetByEmail_NotFound(t *testing.T) {
	repo := newTestDB(t)
	profileRepo := postgresrepo.NewUserProfileRepo(repo)
	cleanupTable(t, repo, "user_profiles")

	_, err := profileRepo.GetByEmail(context.Background(), "missing@example.com")
	if !errors.Is(err, postgresrepo.ErrUserProfileNotFound) {
		t.Fatalf("GetByEmail() error = %v, want %v", err, postgresrepo.ErrUserProfileNotFound)
	}
}

func TestUserProfileRepo_GetByIdentityID(t *testing.T) {
	repo := newTestDB(t)
	profileRepo := postgresrepo.NewUserProfileRepo(repo)
	cleanupTable(t, repo, "user_profiles")

	identityID := uuid.New()
	now := time.Now().UTC()
	profile := &records.UserProfile{
		ID:         uuid.New(),
		IdentityID: identityID,
		Email:      "identity@example.com",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := profileRepo.Create(context.Background(), profile); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := profileRepo.GetByIdentityID(context.Background(), identityID)
	if err != nil {
		t.Fatalf("GetByIdentityID() error = %v", err)
	}

	if got.ID != profile.ID {
		t.Errorf("ID = %v, want %v", got.ID, profile.ID)
	}
}

func TestUserProfileRepo_GetByIdentityID_NotFound(t *testing.T) {
	repo := newTestDB(t)
	profileRepo := postgresrepo.NewUserProfileRepo(repo)
	cleanupTable(t, repo, "user_profiles")

	_, err := profileRepo.GetByIdentityID(context.Background(), uuid.New())
	if !errors.Is(err, postgresrepo.ErrUserProfileNotFound) {
		t.Fatalf("GetByIdentityID() error = %v, want %v", err, postgresrepo.ErrUserProfileNotFound)
	}
}

func TestUserProfileRepo_Update(t *testing.T) {
	repo := newTestDB(t)
	profileRepo := postgresrepo.NewUserProfileRepo(repo)
	cleanupTable(t, repo, "user_profiles")

	now := time.Now().UTC()
	profile := &records.UserProfile{
		ID:        uuid.New(),
		IdentityID: uuid.New(),
		Email:     "update@example.com",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := profileRepo.Create(context.Background(), profile); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	profile.DisplayName = "Updated Name"
	profile.BIO = "new bio"
	profile.AvatarURL = "https://example.com/new.png"
	profile.UpdatedAt = time.Now().UTC()

	if err := profileRepo.Update(context.Background(), profile); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	got, err := profileRepo.GetByID(context.Background(), profile.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.DisplayName != "Updated Name" {
		t.Errorf("DisplayName = %v, want Updated Name", got.DisplayName)
	}
	if got.BIO != "new bio" {
		t.Errorf("BIO = %v, want new bio", got.BIO)
	}
	if got.AvatarURL != "https://example.com/new.png" {
		t.Errorf("AvatarURL = %v, want https://example.com/new.png", got.AvatarURL)
	}
}

func TestUserProfileRepo_Update_NotFound(t *testing.T) {
	repo := newTestDB(t)
	profileRepo := postgresrepo.NewUserProfileRepo(repo)
	cleanupTable(t, repo, "user_profiles")

	profile := &records.UserProfile{
		ID:        uuid.New(),
		IdentityID: uuid.New(),
		Email:     "missing@example.com",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	err := profileRepo.Update(context.Background(), profile)
	if !errors.Is(err, postgresrepo.ErrUserProfileNotFound) {
		t.Fatalf("Update() error = %v, want %v", err, postgresrepo.ErrUserProfileNotFound)
	}
}

func TestUserProfileRepo_Delete(t *testing.T) {
	repo := newTestDB(t)
	profileRepo := postgresrepo.NewUserProfileRepo(repo)
	cleanupTable(t, repo, "user_profiles")

	now := time.Now().UTC()
	profile := &records.UserProfile{
		ID:        uuid.New(),
		IdentityID: uuid.New(),
		Email:     "delete@example.com",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := profileRepo.Create(context.Background(), profile); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := profileRepo.Delete(context.Background(), profile.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err := profileRepo.GetByID(context.Background(), profile.ID)
	if !errors.Is(err, postgresrepo.ErrUserProfileNotFound) {
		t.Fatalf("GetByID() after Delete() error = %v, want %v", err, postgresrepo.ErrUserProfileNotFound)
	}
}

func TestUserProfileRepo_Delete_NotFound(t *testing.T) {
	repo := newTestDB(t)
	profileRepo := postgresrepo.NewUserProfileRepo(repo)
	cleanupTable(t, repo, "user_profiles")

	err := profileRepo.Delete(context.Background(), uuid.New())
	if !errors.Is(err, postgresrepo.ErrUserProfileNotFound) {
		t.Fatalf("Delete() error = %v, want %v", err, postgresrepo.ErrUserProfileNotFound)
	}
}
