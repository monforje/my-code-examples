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

func TestAuthEventRepo_Create(t *testing.T) {
	repo := newTestDB(t)
	eventRepo := postgres.NewAuthEventRepo(repo)
	cleanupTable(t, repo, "auth_events")

	event := &records.AuthEvent{
		ID:        uuid.New(),
		EventType: "login",
		IPAddress: &net.IP{192, 168, 1, 1},
		UserAgent: "test-agent",
		Metadata:  []byte(`{"browser":"chrome"}`),
		CreatedAt: time.Now(),
	}

	err := eventRepo.Create(context.Background(), event)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := eventRepo.GetByID(context.Background(), event.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.EventType != "login" {
		t.Errorf("EventType = %v, want login", got.EventType)
	}
}

func TestAuthEventRepo_GetByIdentityID(t *testing.T) {
	repo := newTestDB(t)
	eventRepo := postgres.NewAuthEventRepo(repo)
	cleanupTable(t, repo, "auth_events")

	identityID := createTestIdentity(t, repo)
	now := time.Now()

	for i := 0; i < 3; i++ {
		event := &records.AuthEvent{
			ID:         uuid.New(),
			IdentityID: &identityID,
			EventType:  "login",
			UserAgent:  "test-agent",
			CreatedAt:  now,
		}
		if err := eventRepo.Create(context.Background(), event); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	events, err := eventRepo.GetByIdentityID(context.Background(), identityID)
	if err != nil {
		t.Fatalf("GetByIdentityID() error = %v", err)
	}

	if len(events) != 3 {
		t.Errorf("len(events) = %d, want 3", len(events))
	}
}

func TestAuthEventRepo_GetByEventType(t *testing.T) {
	repo := newTestDB(t)
	eventRepo := postgres.NewAuthEventRepo(repo)
	cleanupTable(t, repo, "auth_events")

	now := time.Now()

	for i := 0; i < 2; i++ {
		event := &records.AuthEvent{
			ID:        uuid.New(),
			EventType: "logout",
			UserAgent: "test-agent",
			CreatedAt: now,
		}
		if err := eventRepo.Create(context.Background(), event); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	events, err := eventRepo.GetByEventType(context.Background(), "logout")
	if err != nil {
		t.Fatalf("GetByEventType() error = %v", err)
	}

	if len(events) != 2 {
		t.Errorf("len(events) = %d, want 2", len(events))
	}
}
