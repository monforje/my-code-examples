package postgresrepo_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"auth/internal/models/records"
	postgres "auth/internal/repository/postgres"
)

func TestVerificationCodeRepo_Create(t *testing.T) {
	repo := newTestDB(t)
	codeRepo := postgres.NewVerificationCodeRepo(repo)
	cleanupTable(t, repo, "verification_codes")

	code := &records.VerificationCode{
		ID:            uuid.New(),
		Email:         strPtr("test@example.com"),
		Purpose:       "register",
		CodeHash:      "123456",
		AttemptsCount: 0,
		MaxAttempts:   5,
		ExpiresAt:     time.Now().Add(15 * time.Minute),
		CreatedAt:     time.Now(),
	}

	err := codeRepo.Create(context.Background(), code)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := codeRepo.GetByID(context.Background(), code.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.Purpose != "register" {
		t.Errorf("Purpose = %v, want register", got.Purpose)
	}
}

func TestVerificationCodeRepo_GetActiveByEmailAndPurpose(t *testing.T) {
	repo := newTestDB(t)
	codeRepo := postgres.NewVerificationCodeRepo(repo)
	cleanupTable(t, repo, "verification_codes")

	code := &records.VerificationCode{
		ID:            uuid.New(),
		Email:         strPtr("find@example.com"),
		Purpose:       "password_forgot",
		CodeHash:      "654321",
		AttemptsCount: 0,
		MaxAttempts:   5,
		ExpiresAt:     time.Now().Add(15 * time.Minute),
		CreatedAt:     time.Now(),
	}

	if err := codeRepo.Create(context.Background(), code); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := codeRepo.GetActiveByEmailAndPurpose(context.Background(), "find@example.com", "password_forgot")
	if err != nil {
		t.Fatalf("GetActiveByEmailAndPurpose() error = %v", err)
	}

	if got.ID != code.ID {
		t.Errorf("ID = %v, want %v", got.ID, code.ID)
	}
}

func TestVerificationCodeRepo_IncrementAttempts(t *testing.T) {
	repo := newTestDB(t)
	codeRepo := postgres.NewVerificationCodeRepo(repo)
	cleanupTable(t, repo, "verification_codes")

	code := &records.VerificationCode{
		ID:            uuid.New(),
		Email:         strPtr("inc@example.com"),
		Purpose:       "register",
		CodeHash:      "111111",
		AttemptsCount: 0,
		MaxAttempts:   5,
		ExpiresAt:     time.Now().Add(15 * time.Minute),
		CreatedAt:     time.Now(),
	}

	if err := codeRepo.Create(context.Background(), code); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := codeRepo.IncrementAttempts(context.Background(), code.ID); err != nil {
		t.Fatalf("IncrementAttempts() error = %v", err)
	}

	got, err := codeRepo.GetByID(context.Background(), code.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.AttemptsCount != 1 {
		t.Errorf("AttemptsCount = %d, want 1", got.AttemptsCount)
	}
}

func TestVerificationCodeRepo_Consume(t *testing.T) {
	repo := newTestDB(t)
	codeRepo := postgres.NewVerificationCodeRepo(repo)
	cleanupTable(t, repo, "verification_codes")

	code := &records.VerificationCode{
		ID:            uuid.New(),
		Email:         strPtr("consume@example.com"),
		Purpose:       "register",
		CodeHash:      "222222",
		AttemptsCount: 0,
		MaxAttempts:   5,
		ExpiresAt:     time.Now().Add(15 * time.Minute),
		CreatedAt:     time.Now(),
	}

	if err := codeRepo.Create(context.Background(), code); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := codeRepo.Consume(context.Background(), code.ID); err != nil {
		t.Fatalf("Consume() error = %v", err)
	}

	got, err := codeRepo.GetByID(context.Background(), code.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.ConsumedAt == nil {
		t.Error("ConsumedAt is nil, want non-nil")
	}
}

func strPtr(s string) *string {
	return &s
}
