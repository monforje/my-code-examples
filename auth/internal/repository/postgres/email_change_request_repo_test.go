package postgresrepo_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"auth/internal/models/records"
	postgres "auth/internal/repository/postgres"
)

func TestEmailChangeRequestRepo_Create(t *testing.T) {
	repo := newTestDB(t)
	reqRepo := postgres.NewEmailChangeRequestRepo(repo)
	cleanupTable(t, repo, "email_change_requests")
	identityID := createTestIdentity(t, repo)

	req := &records.EmailChangeRequest{
		ID:         uuid.New(),
		IdentityID: identityID,
		NewEmail:   "new@example.com",
		Status:     "pending",
		ExpiresAt:  time.Now().Add(15 * time.Minute),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err := reqRepo.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := reqRepo.GetByID(context.Background(), req.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.NewEmail != "new@example.com" {
		t.Errorf("NewEmail = %v, want new@example.com", got.NewEmail)
	}
}

func TestEmailChangeRequestRepo_GetActiveByIdentityID(t *testing.T) {
	repo := newTestDB(t)
	reqRepo := postgres.NewEmailChangeRequestRepo(repo)
	cleanupTable(t, repo, "email_change_requests")

	identityID := createTestIdentity(t, repo)
	req := &records.EmailChangeRequest{
		ID:         uuid.New(),
		IdentityID: identityID,
		NewEmail:   "active@example.com",
		Status:     "pending",
		ExpiresAt:  time.Now().Add(15 * time.Minute),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := reqRepo.Create(context.Background(), req); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := reqRepo.GetActiveByIdentityIDAndStatus(context.Background(), identityID, "pending")
	if err != nil {
		t.Fatalf("GetActiveByIdentityID() error = %v", err)
	}

	if got.ID != req.ID {
		t.Errorf("ID = %v, want %v", got.ID, req.ID)
	}
}

func TestEmailChangeRequestRepo_SetStatus(t *testing.T) {
	repo := newTestDB(t)
	reqRepo := postgres.NewEmailChangeRequestRepo(repo)
	cleanupTable(t, repo, "email_change_requests")
	identityID := createTestIdentity(t, repo)

	req := &records.EmailChangeRequest{
		ID:         uuid.New(),
		IdentityID: identityID,
		NewEmail:   "status@example.com",
		Status:     "pending",
		ExpiresAt:  time.Now().Add(15 * time.Minute),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := reqRepo.Create(context.Background(), req); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := reqRepo.SetStatus(context.Background(), req.ID, "verified"); err != nil {
		t.Fatalf("SetStatus() error = %v", err)
	}

	got, err := reqRepo.GetByID(context.Background(), req.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.Status != "verified" {
		t.Errorf("Status = %v, want verified", got.Status)
	}
}
