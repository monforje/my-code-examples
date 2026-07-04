// Package events
package events

import "time"

type EventType string

const (
	EventIdentityCreated EventType = "identity.created"
	EventIdentityUpdated EventType = "identity.updated"
	EventIdentityDeleted EventType = "identity.deleted"
)

type Event struct {
	ID         string    `json:"id"`
	Type       EventType `json:"type"`
	OccurredAt time.Time `json:"occurred_at"`
	Data       any       `json:"data"`
}

func NewEvent(eventType EventType, data any) Event {
	now := time.Now().UTC()
	return Event{
		ID:         now.Format("20060102150405.000000000"),
		Type:       eventType,
		OccurredAt: now,
		Data:       data,
	}
}
