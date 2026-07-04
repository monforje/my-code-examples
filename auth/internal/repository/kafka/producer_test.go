package kafkarepo_test

import (
	kafkarepo "auth/internal/repository/kafka"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/twmb/franz-go/pkg/kgo"
)

type mockKafkaClient struct {
	records []*kgo.Record
	err     error
}

func (m *mockKafkaClient) ProduceSync(_ context.Context, rs ...*kgo.Record) kgo.ProduceResults {
	m.records = append(m.records, rs...)
	results := make(kgo.ProduceResults, 0, len(rs))
	for _, record := range rs {
		results = append(results, kgo.ProduceResult{Record: record, Err: m.err})
	}
	return results
}

func TestProducer_PublishIdentityCreated(t *testing.T) {
	client := &mockKafkaClient{}
	producer := kafkarepo.NewProducer(client)

	err := producer.PublishIdentityCreated(context.Background(), kafkarepo.IdentityCreatedPayload{
		IdentityID: "identity-1",
		Email:      "user@example.com",
	})
	if err != nil {
		t.Fatalf("PublishIdentityCreated() error = %v", err)
	}

	record := onlyRecord(t, client)
	if record.Topic != string(kafkarepo.EventIdentityCreated) {
		t.Fatalf("Topic = %s, want %s", record.Topic, kafkarepo.EventIdentityCreated)
	}
	if string(record.Key) != "identity-1" {
		t.Fatalf("Key = %s, want identity-1", string(record.Key))
	}

	event := decodeEvent(t, record.Value)
	if event.Type != kafkarepo.EventIdentityCreated {
		t.Fatalf("Type = %s, want %s", event.Type, kafkarepo.EventIdentityCreated)
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
	client := &mockKafkaClient{}
	producer := kafkarepo.NewProducer(client)
	verified := true

	err := producer.PublishIdentityUpdated(context.Background(), kafkarepo.IdentityUpdatedPayload{
		IdentityID:    "identity-1",
		Email:         "new@example.com",
		Status:        "active",
		EmailVerified: &verified,
	})
	if err != nil {
		t.Fatalf("PublishIdentityUpdated() error = %v", err)
	}

	assertTopicAndType(t, onlyRecord(t, client), kafkarepo.EventIdentityUpdated)
}

func TestProducer_PublishIdentityDeleted(t *testing.T) {
	client := &mockKafkaClient{}
	producer := kafkarepo.NewProducer(client)

	err := producer.PublishIdentityDeleted(context.Background(), kafkarepo.IdentityDeletedPayload{IdentityID: "identity-1"})
	if err != nil {
		t.Fatalf("PublishIdentityDeleted() error = %v", err)
	}

	assertTopicAndType(t, onlyRecord(t, client), kafkarepo.EventIdentityDeleted)
}

func TestProducer_PublishIdentityLogin(t *testing.T) {
	client := &mockKafkaClient{}
	producer := kafkarepo.NewProducer(client)

	err := producer.PublishIdentityLogin(context.Background(), kafkarepo.IdentityLoginPayload{IdentityID: "identity-1", Email: "user@example.com"})
	if err != nil {
		t.Fatalf("PublishIdentityLogin() error = %v", err)
	}

	assertTopicAndType(t, onlyRecord(t, client), kafkarepo.EventIdentityLogin)
}

func TestProducer_PublishIdentityLogout(t *testing.T) {
	client := &mockKafkaClient{}
	producer := kafkarepo.NewProducer(client)

	err := producer.PublishIdentityLogout(context.Background(), kafkarepo.IdentityLogoutPayload{IdentityID: "identity-1"})
	if err != nil {
		t.Fatalf("PublishIdentityLogout() error = %v", err)
	}

	assertTopicAndType(t, onlyRecord(t, client), kafkarepo.EventIdentityLogout)
}

func TestProducer_PublishNotificationCodeEvents(t *testing.T) {
	tests := []struct {
		name      string
		wantEvent kafkarepo.EventType
		publish   func(*kafkarepo.Producer) error
	}{
		{
			name:      "verification code",
			wantEvent: kafkarepo.EventVerificationCodeSend,
			publish: func(p *kafkarepo.Producer) error {
				return p.PublishVerificationCodeSend(context.Background(), kafkarepo.VerificationCodeSendPayload{Email: "user@example.com", Code: "111111", Purpose: "register"})
			},
		},
		{
			name:      "password reset code",
			wantEvent: kafkarepo.EventPasswordResetCodeSend,
			publish: func(p *kafkarepo.Producer) error {
				return p.PublishPasswordResetCodeSend(context.Background(), kafkarepo.VerificationCodeSendPayload{Email: "user@example.com", Code: "111111", Purpose: "password_forgot"})
			},
		},
		{
			name:      "password change code",
			wantEvent: kafkarepo.EventPasswordChangeCodeSend,
			publish: func(p *kafkarepo.Producer) error {
				return p.PublishPasswordChangeCodeSend(context.Background(), kafkarepo.VerificationCodeSendPayload{Email: "user@example.com", Code: "111111", Purpose: "password_change"})
			},
		},
		{
			name:      "email change code",
			wantEvent: kafkarepo.EventEmailChangeCodeSend,
			publish: func(p *kafkarepo.Producer) error {
				return p.PublishEmailChangeCodeSend(context.Background(), kafkarepo.VerificationCodeSendPayload{Email: "new@example.com", Code: "111111", Purpose: "email_change"})
			},
		},
		{
			name:      "account delete code",
			wantEvent: kafkarepo.EventAccountDeleteCodeSend,
			publish: func(p *kafkarepo.Producer) error {
				return p.PublishAccountDeleteCodeSend(context.Background(), kafkarepo.VerificationCodeSendPayload{Email: "user@example.com", Code: "111111", Purpose: "account_delete"})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &mockKafkaClient{}
			producer := kafkarepo.NewProducer(client)

			if err := tt.publish(producer); err != nil {
				t.Fatalf("publish() error = %v", err)
			}

			record := onlyRecord(t, client)
			assertTopicAndType(t, record, tt.wantEvent)
			if string(record.Key) == "" {
				t.Fatal("Key is empty")
			}
		})
	}
}

func TestProducer_PublishReturnsProduceError(t *testing.T) {
	wantErr := errors.New("produce failed")
	client := &mockKafkaClient{err: wantErr}
	producer := kafkarepo.NewProducer(client)

	err := producer.PublishIdentityCreated(context.Background(), kafkarepo.IdentityCreatedPayload{IdentityID: "identity-1", Email: "user@example.com"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want %v", err, wantErr)
	}
}

func onlyRecord(t *testing.T, client *mockKafkaClient) *kgo.Record {
	t.Helper()
	if len(client.records) != 1 {
		t.Fatalf("records count = %d, want 1", len(client.records))
	}
	return client.records[0]
}

func assertTopicAndType(t *testing.T, record *kgo.Record, want kafkarepo.EventType) {
	t.Helper()
	if record.Topic != string(want) {
		t.Fatalf("Topic = %s, want %s", record.Topic, want)
	}
	event := decodeEvent(t, record.Value)
	if event.Type != want {
		t.Fatalf("Type = %s, want %s", event.Type, want)
	}
}

func decodeEvent(t *testing.T, value []byte) kafkarepo.Event {
	t.Helper()
	var event kafkarepo.Event
	if err := json.Unmarshal(value, &event); err != nil {
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
