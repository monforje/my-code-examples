package postgresrepo_test

import (
	"context"
	"testing"
	"time"

	"auth/internal/models/records"
	postgres "auth/internal/repository/postgres"
)

func TestCredentialRepo_Create(t *testing.T) {
	repo := newTestDB(t)
	credentialRepo := postgres.NewCredentialRepo(repo)
	cleanupTable(t, repo, "credentials")
	identityID := createTestIdentity(t, repo)

	credential := &records.Credential{
		IdentityID:        identityID,
		PasswordHash:      "hashed_password",
		PasswordChangedAt: time.Now(),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	err := credentialRepo.Create(context.Background(), credential)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := credentialRepo.GetByIdentityID(context.Background(), credential.IdentityID)
	if err != nil {
		t.Fatalf("GetByIdentityID() error = %v", err)
	}

	if got.PasswordHash != credential.PasswordHash {
		t.Errorf("PasswordHash = %v, want %v", got.PasswordHash, credential.PasswordHash)
	}
}

func TestCredentialRepo_UpdatePassword(t *testing.T) {
	repo := newTestDB(t)
	credentialRepo := postgres.NewCredentialRepo(repo)
	cleanupTable(t, repo, "credentials")

	identityID := createTestIdentity(t, repo)
	credential := &records.Credential{
		IdentityID:        identityID,
		PasswordHash:      "old_hash",
		PasswordChangedAt: time.Now(),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := credentialRepo.Create(context.Background(), credential); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := credentialRepo.UpdatePassword(context.Background(), identityID, "new_hash"); err != nil {
		t.Fatalf("UpdatePassword() error = %v", err)
	}

	got, err := credentialRepo.GetByIdentityID(context.Background(), identityID)
	if err != nil {
		t.Fatalf("GetByIdentityID() error = %v", err)
	}

	if got.PasswordHash != "new_hash" {
		t.Errorf("PasswordHash = %v, want new_hash", got.PasswordHash)
	}
}
