package postgresrepo_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/google/uuid"

	"auth/internal/models/records"
	postgres "auth/internal/repository/postgres"
)

func TestSessionRepo_Create(t *testing.T) {
	repo := newTestDB(t)
	sessionRepo := postgres.NewSessionRepo(repo)
	cleanupTable(t, repo, "sessions")
	identityID := createTestIdentity(t, repo)

	session := &records.Session{
		ID:               uuid.New(),
		IdentityID:       identityID,
		RefreshTokenHash: "token_hash",
		UserAgent:        "test-agent",
		IPAddress:        &net.IP{192, 168, 1, 1},
		ExpiresAt:        time.Now().Add(24 * time.Hour),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	err := sessionRepo.Create(context.Background(), session)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := sessionRepo.GetByID(context.Background(), session.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.RefreshTokenHash != session.RefreshTokenHash {
		t.Errorf("RefreshTokenHash = %v, want %v", got.RefreshTokenHash, session.RefreshTokenHash)
	}
}

func TestSessionRepo_GetByRefreshTokenHash(t *testing.T) {
	repo := newTestDB(t)
	sessionRepo := postgres.NewSessionRepo(repo)
	cleanupTable(t, repo, "sessions")
	identityID := createTestIdentity(t, repo)

	session := &records.Session{
		ID:               uuid.New(),
		IdentityID:       identityID,
		RefreshTokenHash: "find_hash",
		ExpiresAt:        time.Now().Add(24 * time.Hour),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := sessionRepo.Create(context.Background(), session); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := sessionRepo.GetByRefreshTokenHash(context.Background(), "find_hash")
	if err != nil {
		t.Fatalf("GetByRefreshTokenHash() error = %v", err)
	}

	if got.ID != session.ID {
		t.Errorf("ID = %v, want %v", got.ID, session.ID)
	}
}

func TestSessionRepo_GetActiveByIdentityID(t *testing.T) {
	repo := newTestDB(t)
	sessionRepo := postgres.NewSessionRepo(repo)
	cleanupTable(t, repo, "sessions")

	identityID := createTestIdentity(t, repo)
	now := time.Now()

	for i := 0; i < 3; i++ {
		session := &records.Session{
			ID:               uuid.New(),
			IdentityID:       identityID,
			RefreshTokenHash: "hash_" + string(rune('a'+i)),
			ExpiresAt:        now.Add(24 * time.Hour),
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		if err := sessionRepo.Create(context.Background(), session); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	sessions, err := sessionRepo.GetActiveByIdentityID(context.Background(), identityID)
	if err != nil {
		t.Fatalf("GetActiveByIdentityID() error = %v", err)
	}

	if len(sessions) != 3 {
		t.Errorf("len(sessions) = %d, want 3", len(sessions))
	}
}

func TestSessionRepo_Revoke(t *testing.T) {
	repo := newTestDB(t)
	sessionRepo := postgres.NewSessionRepo(repo)
	cleanupTable(t, repo, "sessions")
	identityID := createTestIdentity(t, repo)

	session := &records.Session{
		ID:               uuid.New(),
		IdentityID:       identityID,
		RefreshTokenHash: "revoke_hash",
		ExpiresAt:        time.Now().Add(24 * time.Hour),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := sessionRepo.Create(context.Background(), session); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := sessionRepo.Revoke(context.Background(), session.ID); err != nil {
		t.Fatalf("Revoke() error = %v", err)
	}

	got, err := sessionRepo.GetByID(context.Background(), session.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.RevokedAt == nil {
		t.Error("RevokedAt is nil, want non-nil")
	}
}

func TestSessionRepo_RevokeAllByIdentityID(t *testing.T) {
	repo := newTestDB(t)
	sessionRepo := postgres.NewSessionRepo(repo)
	cleanupTable(t, repo, "sessions")

	identityID := createTestIdentity(t, repo)
	now := time.Now()

	for i := 0; i < 3; i++ {
		session := &records.Session{
			ID:               uuid.New(),
			IdentityID:       identityID,
			RefreshTokenHash: "hash_" + string(rune('a'+i)),
			ExpiresAt:        now.Add(24 * time.Hour),
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		if err := sessionRepo.Create(context.Background(), session); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	if err := sessionRepo.RevokeAllByIdentityID(context.Background(), identityID); err != nil {
		t.Fatalf("RevokeAllByIdentityID() error = %v", err)
	}

	sessions, err := sessionRepo.GetActiveByIdentityID(context.Background(), identityID)
	if err != nil {
		t.Fatalf("GetActiveByIdentityID() error = %v", err)
	}

	if len(sessions) != 0 {
		t.Errorf("len(sessions) = %d, want 0", len(sessions))
	}
}
