package records

import (
	"time"

	"github.com/google/uuid"
)

type ProcessedEvent struct {
	EventID     string    `db:"event_id"`
	EventType   string    `db:"event_type"`
	AggregateID uuid.UUID `db:"aggregate_id"`
	ProcessedAt time.Time `db:"processed_at"`
}
