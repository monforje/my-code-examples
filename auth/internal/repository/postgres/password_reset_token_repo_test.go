package postgresrepo_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"auth/internal/models/records"
	postgres "auth/internal/repository/postgres"
)

func TestPasswordResetTokenRepo_Create(t *testing.T) {
	repo := newTestDB(t)
	tokenRepo := postgres.NewPasswordResetTokenRepo(repo)
	cleanupTable(t, repo, "password_reset_tokens")
	identityID := createTestIdentity(t, repo)

	token := &records.PasswordResetToken{
		ID:         uuid.New(),
		IdentityID: identityID,
		TokenHash:  "reset_hash",
		ExpiresAt:  time.Now().Add(1 * time.Hour),
		CreatedAt:  time.Now(),
	}

	err := tokenRepo.Create(context.Background(), token)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := tokenRepo.GetByID(context.Background(), token.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.TokenHash != "reset_hash" {
		t.Errorf("TokenHash = %v, want reset_hash", got.TokenHash)
	}
}

func TestPasswordResetTokenRepo_GetByTokenHash(t *testing.T) {
	repo := newTestDB(t)
	tokenRepo := postgres.NewPasswordResetTokenRepo(repo)
	cleanupTable(t, repo, "password_reset_tokens")
	identityID := createTestIdentity(t, repo)

	token := &records.PasswordResetToken{
		ID:         uuid.New(),
		IdentityID: identityID,
		TokenHash:  "find_hash",
		ExpiresAt:  time.Now().Add(1 * time.Hour),
		CreatedAt:  time.Now(),
	}

	if err := tokenRepo.Create(context.Background(), token); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := tokenRepo.GetByTokenHash(context.Background(), "find_hash")
	if err != nil {
		t.Fatalf("GetByTokenHash() error = %v", err)
	}

	if got.ID != token.ID {
		t.Errorf("ID = %v, want %v", got.ID, token.ID)
	}
}

func TestPasswordResetTokenRepo_Consume(t *testing.T) {
	repo := newTestDB(t)
	tokenRepo := postgres.NewPasswordResetTokenRepo(repo)
	cleanupTable(t, repo, "password_reset_tokens")
	identityID := createTestIdentity(t, repo)

	token := &records.PasswordResetToken{
		ID:         uuid.New(),
		IdentityID: identityID,
		TokenHash:  "consume_hash",
		ExpiresAt:  time.Now().Add(1 * time.Hour),
		CreatedAt:  time.Now(),
	}

	if err := tokenRepo.Create(context.Background(), token); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := tokenRepo.Consume(context.Background(), token.ID); err != nil {
		t.Fatalf("Consume() error = %v", err)
	}

	got, err := tokenRepo.GetByID(context.Background(), token.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.ConsumedAt == nil {
		t.Error("ConsumedAt is nil, want non-nil")
	}
}
