package postgresrepo_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"auth/internal/models/records"
	postgres "auth/internal/repository/postgres"
)

func TestIdentityRepo_Create(t *testing.T) {
	repo := newTestDB(t)
	identityRepo := postgres.NewIdentityRepo(repo)
	cleanupTable(t, repo, "identities")

	identity := &records.Identity{
		ID:        uuid.New(),
		Email:     "test@example.com",
		Status:    "pending_verification",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := identityRepo.Create(context.Background(), identity)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := identityRepo.GetByID(context.Background(), identity.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.Email != identity.Email {
		t.Errorf("Email = %v, want %v", got.Email, identity.Email)
	}
}

func TestIdentityRepo_GetByEmail(t *testing.T) {
	repo := newTestDB(t)
	identityRepo := postgres.NewIdentityRepo(repo)
	cleanupTable(t, repo, "identities")

	identity := &records.Identity{
		ID:        uuid.New(),
		Email:     "find@example.com",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := identityRepo.Create(context.Background(), identity); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := identityRepo.GetByEmail(context.Background(), "find@example.com")
	if err != nil {
		t.Fatalf("GetByEmail() error = %v", err)
	}

	if got.ID != identity.ID {
		t.Errorf("ID = %v, want %v", got.ID, identity.ID)
	}
}

func TestIdentityRepo_Update(t *testing.T) {
	repo := newTestDB(t)
	identityRepo := postgres.NewIdentityRepo(repo)
	cleanupTable(t, repo, "identities")

	identity := &records.Identity{
		ID:        uuid.New(),
		Email:     "old@example.com",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := identityRepo.Create(context.Background(), identity); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	identity.Email = "new@example.com"
	identity.UpdatedAt = time.Now()

	if err := identityRepo.Update(context.Background(), identity); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	got, err := identityRepo.GetByID(context.Background(), identity.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.Email != "new@example.com" {
		t.Errorf("Email = %v, want new@example.com", got.Email)
	}
}

func TestIdentityRepo_SetEmailVerified(t *testing.T) {
	repo := newTestDB(t)
	identityRepo := postgres.NewIdentityRepo(repo)
	cleanupTable(t, repo, "identities")

	identity := &records.Identity{
		ID:        uuid.New(),
		Email:     "verify@example.com",
		Status:    "pending_verification",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := identityRepo.Create(context.Background(), identity); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := identityRepo.SetEmailVerified(context.Background(), identity.ID); err != nil {
		t.Fatalf("SetEmailVerified() error = %v", err)
	}

	got, err := identityRepo.GetByID(context.Background(), identity.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if !got.EmailVerified {
		t.Error("EmailVerified = false, want true")
	}
}

func TestIdentityRepo_SetStatus(t *testing.T) {
	repo := newTestDB(t)
	identityRepo := postgres.NewIdentityRepo(repo)
	cleanupTable(t, repo, "identities")

	identity := &records.Identity{
		ID:        uuid.New(),
		Email:     "status@example.com",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := identityRepo.Create(context.Background(), identity); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := identityRepo.SetStatus(context.Background(), identity.ID, "blocked"); err != nil {
		t.Fatalf("SetStatus() error = %v", err)
	}

	got, err := identityRepo.GetByID(context.Background(), identity.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.Status != "blocked" {
		t.Errorf("Status = %v, want blocked", got.Status)
	}
}

func TestIdentityRepo_SoftDelete(t *testing.T) {
	repo := newTestDB(t)
	identityRepo := postgres.NewIdentityRepo(repo)
	cleanupTable(t, repo, "identities")

	identity := &records.Identity{
		ID:        uuid.New(),
		Email:     "delete@example.com",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := identityRepo.Create(context.Background(), identity); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := identityRepo.SoftDelete(context.Background(), identity.ID); err != nil {
		t.Fatalf("SoftDelete() error = %v", err)
	}

	got, err := identityRepo.GetByID(context.Background(), identity.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.DeletedAt == nil {
		t.Error("DeletedAt is nil, want non-nil")
	}
}
