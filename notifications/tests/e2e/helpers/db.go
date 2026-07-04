package e2e_test_helpers

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ProcessedEventRow struct {
	EventID     string
	EventType   string
	AggregateID string
	ProcessedAt time.Time
}

func GetProcessedEventByEventID(t *testing.T, pool *pgxpool.Pool, eventID string) ProcessedEventRow {
	t.Helper()
	var row ProcessedEventRow
	err := pool.QueryRow(context.Background(),
		`SELECT event_id, event_type, aggregate_id, processed_at
		 FROM processed_events WHERE event_id = $1`, eventID,
	).Scan(&row.EventID, &row.EventType, &row.AggregateID, &row.ProcessedAt)
	if err != nil {
		t.Fatalf("GetProcessedEventByEventID(%s): %v", eventID, err)
	}
	return row
}

func CountProcessedEvents(t *testing.T, pool *pgxpool.Pool) int {
	t.Helper()
	var count int
	err := pool.QueryRow(context.Background(), `SELECT count(*) FROM processed_events`).Scan(&count)
	if err != nil {
		t.Fatalf("CountProcessedEvents(): %v", err)
	}
	return count
}
