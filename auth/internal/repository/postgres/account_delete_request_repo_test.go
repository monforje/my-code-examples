package postgresrepo_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"auth/internal/models/records"
	postgres "auth/internal/repository/postgres"
)

func TestAccountDeleteRequestRepo_Create(t *testing.T) {
	repo := newTestDB(t)
	reqRepo := postgres.NewAccountDeleteRequestRepo(repo)
	cleanupTable(t, repo, "account_delete_requests")
	identityID := createTestIdentity(t, repo)

	req := &records.AccountDeleteRequest{
		ID:         uuid.New(),
		IdentityID: identityID,
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

	if got.Status != "pending" {
		t.Errorf("Status = %v, want pending", got.Status)
	}
}

func TestAccountDeleteRequestRepo_GetActiveByIdentityID(t *testing.T) {
	repo := newTestDB(t)
	reqRepo := postgres.NewAccountDeleteRequestRepo(repo)
	cleanupTable(t, repo, "account_delete_requests")

	identityID := createTestIdentity(t, repo)
	req := &records.AccountDeleteRequest{
		ID:         uuid.New(),
		IdentityID: identityID,
		Status:     "pending",
		ExpiresAt:  time.Now().Add(15 * time.Minute),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := reqRepo.Create(context.Background(), req); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := reqRepo.GetActiveByIdentityID(context.Background(), identityID)
	if err != nil {
		t.Fatalf("GetActiveByIdentityID() error = %v", err)
	}

	if got.ID != req.ID {
		t.Errorf("ID = %v, want %v", got.ID, req.ID)
	}
}

func TestAccountDeleteRequestRepo_SetStatus(t *testing.T) {
	repo := newTestDB(t)
	reqRepo := postgres.NewAccountDeleteRequestRepo(repo)
	cleanupTable(t, repo, "account_delete_requests")
	identityID := createTestIdentity(t, repo)

	req := &records.AccountDeleteRequest{
		ID:         uuid.New(),
		IdentityID: identityID,
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
