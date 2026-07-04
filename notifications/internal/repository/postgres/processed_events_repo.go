package postgresrepo

import (
	"context"
	"errors"

	"notifications/internal/models/records"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrProcessedEventNotFound = errors.New("processed event not found")

type ProcessedEventsRepo struct {
	*Repo
}

func NewProcessedEventsRepo(repo *Repo) *ProcessedEventsRepo {
	return &ProcessedEventsRepo{Repo: repo}
}

func (r *ProcessedEventsRepo) Create(ctx context.Context, event *records.ProcessedEvent) error {
	_, err := r.Exec(ctx, `
		INSERT INTO processed_events (event_id, event_type, aggregate_id, processed_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (event_id) DO NOTHING
	`, event.EventID, event.EventType, event.AggregateID, event.ProcessedAt)
	return err
}

func (r *ProcessedEventsRepo) GetByEventID(ctx context.Context, eventID string) (*records.ProcessedEvent, error) {
	var e records.ProcessedEvent
	err := r.QueryRow(ctx, `
		SELECT event_id, event_type, aggregate_id, processed_at
		FROM processed_events
		WHERE event_id = $1
	`, eventID).Scan(&e.EventID, &e.EventType, &e.AggregateID, &e.ProcessedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrProcessedEventNotFound
	}
	return &e, err
}

func (r *ProcessedEventsRepo) GetByAggregateID(ctx context.Context, aggregateID uuid.UUID) ([]records.ProcessedEvent, error) {
	rows, err := r.Query(ctx, `
		SELECT event_id, event_type, aggregate_id, processed_at
		FROM processed_events
		WHERE aggregate_id = $1
		ORDER BY processed_at DESC
	`, aggregateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []records.ProcessedEvent
	for rows.Next() {
		var e records.ProcessedEvent
		if err := rows.Scan(&e.EventID, &e.EventType, &e.AggregateID, &e.ProcessedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

func (r *ProcessedEventsRepo) Update(ctx context.Context, event *records.ProcessedEvent) error {
	tag, err := r.Exec(ctx, `
		UPDATE processed_events
		SET event_type = $2, aggregate_id = $3, processed_at = $4
		WHERE event_id = $1
	`, event.EventID, event.EventType, event.AggregateID, event.ProcessedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrProcessedEventNotFound
	}
	return nil
}

func (r *ProcessedEventsRepo) Delete(ctx context.Context, eventID string) error {
	tag, err := r.Exec(ctx, `
		DELETE FROM processed_events
		WHERE event_id = $1
	`, eventID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrProcessedEventNotFound
	}
	return nil
}
