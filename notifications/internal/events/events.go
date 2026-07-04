// Package events
package events

import "time"

type EventType string

const (
	EventVerificationCodeSend   EventType = "notification.email.verification_code.send"
	EventPasswordResetCodeSend  EventType = "notification.email.password_reset_code.send"
	EventPasswordChangeCodeSend EventType = "notification.email.password_change_code.send"
	EventEmailChangeCodeSend    EventType = "notification.email.email_change_code.send"
	EventAccountDeleteCodeSend  EventType = "notification.email.account_delete_code.send"
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
