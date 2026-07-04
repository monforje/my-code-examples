package natsrepo_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"auth/internal/events"
	natsrepo "auth/internal/repository/nats"

	"github.com/nats-io/nats.go"
)

type mockNatsClient struct {
	msgs  []*nats.Msg
	err   error
}

func (m *mockNatsClient) PublishMsg(msg *nats.Msg) error {
	m.msgs = append(m.msgs, msg)
	return m.err
}

func TestProducer_PublishIdentityCreated(t *testing.T) {
	client := &mockNatsClient{}
	producer := natsrepo.NewProducer(client)

	err := producer.PublishIdentityCreated(context.Background(), events.IdentityCreatedPayload{
		IdentityID: "identity-1",
		Email:      "user@example.com",
	})
	if err != nil {
		t.Fatalf("PublishIdentityCreated() error = %v", err)
	}

	msg := onlyMsg(t, client)
	if msg.Subject != string(events.EventIdentityCreated) {
		t.Fatalf("Subject = %s, want %s", msg.Subject, events.EventIdentityCreated)
	}
	if msg.Header.Get("X-Key") != "identity-1" {
		t.Fatalf("X-Key = %s, want identity-1", msg.Header.Get("X-Key"))
	}

	event := decodeEvent(t, msg.Data)
	if event.Type != events.EventIdentityCreated {
		t.Fatalf("Type = %s, want %s", event.Type, events.EventIdentityCreated)
	}

	data := event.Data.(map[string]any)
	if data["identity_id"] != "identity-1" {
		t.Fatalf("identity_id = %v, want identity-1", data["identity_id"])
	}
	if data["email"] != "user@example.com" {
		t.Fatalf("email = %v, want user@example.com", data["email"])
	}
}

func TestProducer_PublishIdentityUpdated(t *testing.T) {
	client := &mockNatsClient{}
	producer := natsrepo.NewProducer(client)
	verified := true

	err := producer.PublishIdentityUpdated(context.Background(), events.IdentityUpdatedPayload{
		IdentityID:    "identity-1",
		Email:         "new@example.com",
		Status:        "active",
		EmailVerified: &verified,
	})
	if err != nil {
		t.Fatalf("PublishIdentityUpdated() error = %v", err)
	}

	assertSubjectAndType(t, onlyMsg(t, client), events.EventIdentityUpdated)
}

func TestProducer_PublishIdentityDeleted(t *testing.T) {
	client := &mockNatsClient{}
	producer := natsrepo.NewProducer(client)

	err := producer.PublishIdentityDeleted(context.Background(), events.IdentityDeletedPayload{IdentityID: "identity-1"})
	if err != nil {
		t.Fatalf("PublishIdentityDeleted() error = %v", err)
	}

	assertSubjectAndType(t, onlyMsg(t, client), events.EventIdentityDeleted)
}

func TestProducer_PublishIdentityLogin(t *testing.T) {
	client := &mockNatsClient{}
	producer := natsrepo.NewProducer(client)

	err := producer.PublishIdentityLogin(context.Background(), events.IdentityLoginPayload{IdentityID: "identity-1", Email: "user@example.com"})
	if err != nil {
		t.Fatalf("PublishIdentityLogin() error = %v", err)
	}

	assertSubjectAndType(t, onlyMsg(t, client), events.EventIdentityLogin)
}

func TestProducer_PublishIdentityLogout(t *testing.T) {
	client := &mockNatsClient{}
	producer := natsrepo.NewProducer(client)

	err := producer.PublishIdentityLogout(context.Background(), events.IdentityLogoutPayload{IdentityID: "identity-1"})
	if err != nil {
		t.Fatalf("PublishIdentityLogout() error = %v", err)
	}

	assertSubjectAndType(t, onlyMsg(t, client), events.EventIdentityLogout)
}

func TestProducer_PublishNotificationCodeEvents(t *testing.T) {
	tests := []struct {
		name      string
		wantEvent events.EventType
		publish   func(*natsrepo.Producer) error
	}{
		{
			name:      "verification code",
			wantEvent: events.EventVerificationCodeSend,
			publish: func(p *natsrepo.Producer) error {
				return p.PublishVerificationCodeSend(context.Background(), events.VerificationCodeSendPayload{Email: "user@example.com", Code: "111111", Purpose: "register"})
			},
		},
		{
			name:      "password reset code",
			wantEvent: events.EventPasswordResetCodeSend,
			publish: func(p *natsrepo.Producer) error {
				return p.PublishPasswordResetCodeSend(context.Background(), events.VerificationCodeSendPayload{Email: "user@example.com", Code: "111111", Purpose: "password_forgot"})
			},
		},
		{
			name:      "password change code",
			wantEvent: events.EventPasswordChangeCodeSend,
			publish: func(p *natsrepo.Producer) error {
				return p.PublishPasswordChangeCodeSend(context.Background(), events.VerificationCodeSendPayload{Email: "user@example.com", Code: "111111", Purpose: "password_change"})
			},
		},
		{
			name:      "email change code",
			wantEvent: events.EventEmailChangeCodeSend,
			publish: func(p *natsrepo.Producer) error {
				return p.PublishEmailChangeCodeSend(context.Background(), events.VerificationCodeSendPayload{Email: "new@example.com", Code: "111111", Purpose: "email_change"})
			},
		},
		{
			name:      "account delete code",
			wantEvent: events.EventAccountDeleteCodeSend,
			publish: func(p *natsrepo.Producer) error {
				return p.PublishAccountDeleteCodeSend(context.Background(), events.VerificationCodeSendPayload{Email: "user@example.com", Code: "111111", Purpose: "account_delete"})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &mockNatsClient{}
			producer := natsrepo.NewProducer(client)

			if err := tt.publish(producer); err != nil {
				t.Fatalf("publish() error = %v", err)
			}

			msg := onlyMsg(t, client)
			assertSubjectAndType(t, msg, tt.wantEvent)
			if msg.Header.Get("X-Key") == "" {
				t.Fatal("X-Key is empty")
			}
		})
	}
}

func TestProducer_PublishReturnsPublishError(t *testing.T) {
	wantErr := errors.New("publish failed")
	client := &mockNatsClient{err: wantErr}
	producer := natsrepo.NewProducer(client)

	err := producer.PublishIdentityCreated(context.Background(), events.IdentityCreatedPayload{IdentityID: "identity-1", Email: "user@example.com"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want %v", err, wantErr)
	}
}

func onlyMsg(t *testing.T, client *mockNatsClient) *nats.Msg {
	t.Helper()
	if len(client.msgs) != 1 {
		t.Fatalf("msgs count = %d, want 1", len(client.msgs))
	}
	return client.msgs[0]
}

func assertSubjectAndType(t *testing.T, msg *nats.Msg, want events.EventType) {
	t.Helper()
	if msg.Subject != string(want) {
		t.Fatalf("Subject = %s, want %s", msg.Subject, want)
	}
	event := decodeEvent(t, msg.Data)
	if event.Type != want {
		t.Fatalf("Type = %s, want %s", event.Type, want)
	}
}

func decodeEvent(t *testing.T, data []byte) events.Event {
	t.Helper()
	var event events.Event
	if err := json.Unmarshal(data, &event); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if event.ID == "" {
		t.Fatal("event.ID is empty")
	}
	if event.OccurredAt.IsZero() {
		t.Fatal("event.OccurredAt is zero")
	}
	return event
}
