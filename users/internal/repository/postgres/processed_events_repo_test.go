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

func TestProcessedEventsRepo_Create(t *testing.T) {
	repo := newTestDB(t)
	eventsRepo := postgresrepo.NewProcessedEventsRepo(repo)
	cleanupTable(t, repo, "processed_events")

	event := &records.ProcessedEvent{
		EventID:     uuid.New().String(),
		EventType:   "identity.created",
		AggregateID: uuid.New(),
		ProcessedAt: time.Now().UTC(),
	}

	err := eventsRepo.Create(context.Background(), event)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := eventsRepo.GetByEventID(context.Background(), event.EventID)
	if err != nil {
		t.Fatalf("GetByEventID() error = %v", err)
	}

	if got.EventType != event.EventType {
		t.Errorf("EventType = %v, want %v", got.EventType, event.EventType)
	}
	if got.AggregateID != event.AggregateID {
		t.Errorf("AggregateID = %v, want %v", got.AggregateID, event.AggregateID)
	}
}

func TestProcessedEventsRepo_Create_Duplicate(t *testing.T) {
	repo := newTestDB(t)
	eventsRepo := postgresrepo.NewProcessedEventsRepo(repo)
	cleanupTable(t, repo, "processed_events")

	event := &records.ProcessedEvent{
		EventID:     uuid.New().String(),
		EventType:   "identity.created",
		AggregateID: uuid.New(),
		ProcessedAt: time.Now().UTC(),
	}

	if err := eventsRepo.Create(context.Background(), event); err != nil {
		t.Fatalf("Create() first error = %v", err)
	}

	err := eventsRepo.Create(context.Background(), event)
	if err != nil {
		t.Fatalf("Create() duplicate should not error, got = %v", err)
	}
}

func TestProcessedEventsRepo_GetByEventID_NotFound(t *testing.T) {
	repo := newTestDB(t)
	eventsRepo := postgresrepo.NewProcessedEventsRepo(repo)
	cleanupTable(t, repo, "processed_events")

	_, err := eventsRepo.GetByEventID(context.Background(), uuid.New().String())
	if !errors.Is(err, postgresrepo.ErrProcessedEventNotFound) {
		t.Fatalf("GetByEventID() error = %v, want %v", err, postgresrepo.ErrProcessedEventNotFound)
	}
}

func TestProcessedEventsRepo_GetByAggregateID(t *testing.T) {
	repo := newTestDB(t)
	eventsRepo := postgresrepo.NewProcessedEventsRepo(repo)
	cleanupTable(t, repo, "processed_events")

	aggregateID := uuid.New()
	now := time.Now().UTC()

	e1 := &records.ProcessedEvent{
		EventID:     uuid.New().String(),
		EventType:   "identity.created",
		AggregateID: aggregateID,
		ProcessedAt: now,
	}
	e2 := &records.ProcessedEvent{
		EventID:     uuid.New().String(),
		EventType:   "profile.update",
		AggregateID: aggregateID,
		ProcessedAt: now.Add(time.Second),
	}

	if err := eventsRepo.Create(context.Background(), e1); err != nil {
		t.Fatalf("Create() e1 error = %v", err)
	}
	if err := eventsRepo.Create(context.Background(), e2); err != nil {
		t.Fatalf("Create() e2 error = %v", err)
	}

	events, err := eventsRepo.GetByAggregateID(context.Background(), aggregateID)
	if err != nil {
		t.Fatalf("GetByAggregateID() error = %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("GetByAggregateID() returned %d events, want 2", len(events))
	}
}

func TestProcessedEventsRepo_GetByAggregateID_Empty(t *testing.T) {
	repo := newTestDB(t)
	eventsRepo := postgresrepo.NewProcessedEventsRepo(repo)
	cleanupTable(t, repo, "processed_events")

	events, err := eventsRepo.GetByAggregateID(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("GetByAggregateID() error = %v", err)
	}

	if len(events) != 0 {
		t.Fatalf("GetByAggregateID() returned %d events, want 0", len(events))
	}
}

func TestProcessedEventsRepo_Update(t *testing.T) {
	repo := newTestDB(t)
	eventsRepo := postgresrepo.NewProcessedEventsRepo(repo)
	cleanupTable(t, repo, "processed_events")

	event := &records.ProcessedEvent{
		EventID:     uuid.New().String(),
		EventType:   "identity.created",
		AggregateID: uuid.New(),
		ProcessedAt: time.Now().UTC(),
	}

	if err := eventsRepo.Create(context.Background(), event); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	event.EventType = "profile.update"
	event.ProcessedAt = time.Now().UTC()

	if err := eventsRepo.Update(context.Background(), event); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	got, err := eventsRepo.GetByEventID(context.Background(), event.EventID)
	if err != nil {
		t.Fatalf("GetByEventID() error = %v", err)
	}

	if got.EventType != "profile.update" {
		t.Errorf("EventType = %v, want profile.update", got.EventType)
	}
}

func TestProcessedEventsRepo_Update_NotFound(t *testing.T) {
	repo := newTestDB(t)
	eventsRepo := postgresrepo.NewProcessedEventsRepo(repo)
	cleanupTable(t, repo, "processed_events")

	event := &records.ProcessedEvent{
		EventID:     uuid.New().String(),
		EventType:   "identity.created",
		AggregateID: uuid.New(),
		ProcessedAt: time.Now().UTC(),
	}

	err := eventsRepo.Update(context.Background(), event)
	if !errors.Is(err, postgresrepo.ErrProcessedEventNotFound) {
		t.Fatalf("Update() error = %v, want %v", err, postgresrepo.ErrProcessedEventNotFound)
	}
}

func TestProcessedEventsRepo_Delete(t *testing.T) {
	repo := newTestDB(t)
	eventsRepo := postgresrepo.NewProcessedEventsRepo(repo)
	cleanupTable(t, repo, "processed_events")

	event := &records.ProcessedEvent{
		EventID:     uuid.New().String(),
		EventType:   "identity.created",
		AggregateID: uuid.New(),
		ProcessedAt: time.Now().UTC(),
	}

	if err := eventsRepo.Create(context.Background(), event); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := eventsRepo.Delete(context.Background(), event.EventID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err := eventsRepo.GetByEventID(context.Background(), event.EventID)
	if !errors.Is(err, postgresrepo.ErrProcessedEventNotFound) {
		t.Fatalf("GetByEventID() after Delete() error = %v, want %v", err, postgresrepo.ErrProcessedEventNotFound)
	}
}

func TestProcessedEventsRepo_Delete_NotFound(t *testing.T) {
	repo := newTestDB(t)
	eventsRepo := postgresrepo.NewProcessedEventsRepo(repo)
	cleanupTable(t, repo, "processed_events")

	err := eventsRepo.Delete(context.Background(), uuid.New().String())
	if !errors.Is(err, postgresrepo.ErrProcessedEventNotFound) {
		t.Fatalf("Delete() error = %v, want %v", err, postgresrepo.ErrProcessedEventNotFound)
	}
}
