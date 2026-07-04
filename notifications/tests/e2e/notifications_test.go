package e2e_test

import (
	"encoding/json"
	"testing"
	"time"

	"notifications/internal/events"
	e2ehelpers "notifications/tests/e2e/helpers"

	"github.com/google/uuid"
)

func publishCodeEvent(t *testing.T, eventType events.EventType, identityID string, email string, code string) string {
	t.Helper()
	payload := events.VerificationCodeSendPayload{
		IdentityID: &identityID,
		Email:      email,
		Code:       code,
		Purpose:    "e2e",
	}
	event := events.NewEvent(eventType, payload)
	event.ID = uuid.New().String()

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}
	if err := natsConn.Publish(string(eventType), data); err != nil {
		t.Fatalf("publish event: %v", err)
	}
	if err := natsConn.Flush(); err != nil {
		t.Fatalf("flush nats: %v", err)
	}

	return event.ID
}

func TestNotificationConsumer_VerificationCodeFlow(t *testing.T) {
	resetE2E(t)

	identityID := uuid.New().String()
	eventID := publishCodeEvent(t, events.EventVerificationCodeSend, identityID, "verify@example.com", "111111")
	waitEmailCount(t, 1)

	sent := sender.last()
	if sent.template != "verification" {
		t.Fatalf("template = %q, want verification", sent.template)
	}
	if sent.params.To != "verify@example.com" {
		t.Fatalf("email = %q, want verify@example.com", sent.params.To)
	}
	if sent.params.Code != "111111" {
		t.Fatalf("code = %q, want 111111", sent.params.Code)
	}

	row := e2ehelpers.GetProcessedEventByEventID(t, e2eEnv.PgPool(), eventID)
	if row.EventType != string(events.EventVerificationCodeSend) {
		t.Fatalf("event_type = %q, want %q", row.EventType, events.EventVerificationCodeSend)
	}
	if row.AggregateID != identityID {
		t.Fatalf("aggregate_id = %q, want %q", row.AggregateID, identityID)
	}
}

func TestNotificationConsumer_AllCodeEventsFlow(t *testing.T) {
	resetE2E(t)

	tests := []struct {
		eventType events.EventType
		template  string
	}{
		{events.EventVerificationCodeSend, "verification"},
		{events.EventPasswordResetCodeSend, "password_reset"},
		{events.EventPasswordChangeCodeSend, "password_change"},
		{events.EventEmailChangeCodeSend, "email_change"},
		{events.EventAccountDeleteCodeSend, "delete_account"},
	}

	for i, tt := range tests {
		identityID := uuid.New().String()
		eventID := publishCodeEvent(t, tt.eventType, identityID, tt.template+"@example.com", "222222")
		waitEmailCount(t, i+1)

		row := e2ehelpers.GetProcessedEventByEventID(t, e2eEnv.PgPool(), eventID)
		if row.EventType != string(tt.eventType) {
			t.Fatalf("event_type = %q, want %q", row.EventType, tt.eventType)
		}
	}
}

func TestNotificationConsumer_IdempotentFlow(t *testing.T) {
	resetE2E(t)

	identityID := uuid.New().String()
	payload := events.VerificationCodeSendPayload{
		IdentityID: &identityID,
		Email:      "idempotent@example.com",
		Code:       "333333",
		Purpose:    "e2e",
	}
	event := events.NewEvent(events.EventVerificationCodeSend, payload)
	event.ID = uuid.New().String()
	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}

	for i := 0; i < 2; i++ {
		if err := natsConn.Publish(string(events.EventVerificationCodeSend), data); err != nil {
			t.Fatalf("publish event: %v", err)
		}
		if err := natsConn.Flush(); err != nil {
			t.Fatalf("flush nats: %v", err)
		}
	}

	waitEmailCount(t, 1)
	time.Sleep(200 * time.Millisecond)
	if sender.count() != 1 {
		t.Fatalf("sent emails = %d, want 1", sender.count())
	}
	if count := e2ehelpers.CountProcessedEvents(t, e2eEnv.PgPool()); count != 1 {
		t.Fatalf("processed_events count = %d, want 1", count)
	}
}
