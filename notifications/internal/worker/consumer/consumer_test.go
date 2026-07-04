package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"
	"time"

	"notifications/internal/config"
	"notifications/internal/events"
	"notifications/internal/models/records"
	postgresrepo "notifications/internal/repository/postgres"
	service "notifications/internal/services"
	"notifications/pkg/logger"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

type mockEmailSender struct {
	verification   []service.SendCodeEmailParams
	passwordReset  []service.SendCodeEmailParams
	passwordChange []service.SendCodeEmailParams
	emailChange    []service.SendCodeEmailParams
	deleteAccount  []service.SendCodeEmailParams
	err            error
}

func (m *mockEmailSender) SendVerificationEmail(_ context.Context, params service.SendCodeEmailParams) error {
	if m.err != nil {
		return m.err
	}
	m.verification = append(m.verification, params)
	return nil
}

func (m *mockEmailSender) SendPasswordResetEmail(_ context.Context, params service.SendCodeEmailParams) error {
	if m.err != nil {
		return m.err
	}
	m.passwordReset = append(m.passwordReset, params)
	return nil
}

func (m *mockEmailSender) SendPasswordChangeEmail(_ context.Context, params service.SendCodeEmailParams) error {
	if m.err != nil {
		return m.err
	}
	m.passwordChange = append(m.passwordChange, params)
	return nil
}

func (m *mockEmailSender) SendEmailChangeEmail(_ context.Context, params service.SendCodeEmailParams) error {
	if m.err != nil {
		return m.err
	}
	m.emailChange = append(m.emailChange, params)
	return nil
}

func (m *mockEmailSender) SendDeleteAccountEmail(_ context.Context, params service.SendCodeEmailParams) error {
	if m.err != nil {
		return m.err
	}
	m.deleteAccount = append(m.deleteAccount, params)
	return nil
}

type mockProcessedEvents struct {
	created  []*records.ProcessedEvent
	existing map[string]*records.ProcessedEvent
}

func (m *mockProcessedEvents) Create(_ context.Context, e *records.ProcessedEvent) error {
	m.created = append(m.created, e)
	return nil
}

func (m *mockProcessedEvents) GetByEventID(_ context.Context, id string) (*records.ProcessedEvent, error) {
	if e, ok := m.existing[id]; ok {
		return e, nil
	}
	return nil, postgresrepo.ErrProcessedEventNotFound
}

func newNotificationService(sender *mockEmailSender) *service.NotificationService {
	return service.NewNotificationService(sender)
}

func marshalPayload(t *testing.T, identityID string) []byte {
	t.Helper()
	data, err := json.Marshal(events.VerificationCodeSendPayload{
		IdentityID: &identityID,
		Email:      "user@example.com",
		Code:       "123456",
		Purpose:    "test",
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	return data
}

func TestHandleVerificationCode_Success(t *testing.T) {
	identityID := uuid.New().String()
	sender := &mockEmailSender{}
	svc := newNotificationService(sender)

	gotID, err := handleVerificationCode(context.Background(), svc, marshalPayload(t, identityID))
	if err != nil {
		t.Fatalf("handleVerificationCode() error = %v", err)
	}
	if gotID != identityID {
		t.Fatalf("handleVerificationCode() = %q, want %q", gotID, identityID)
	}
	if len(sender.verification) != 1 {
		t.Fatalf("SendVerificationEmail() called %d times, want 1", len(sender.verification))
	}
}

func TestHandleVerificationCode_MissingIdentityID(t *testing.T) {
	svc := newNotificationService(&mockEmailSender{})
	data, _ := json.Marshal(events.VerificationCodeSendPayload{Email: "user@example.com", Code: "123456"})

	_, err := handleVerificationCode(context.Background(), svc, data)
	if err == nil {
		t.Fatal("handleVerificationCode() error = nil, want error")
	}
}

func TestHandlePasswordResetCode_Success(t *testing.T) {
	identityID := uuid.New().String()
	sender := &mockEmailSender{}
	svc := newNotificationService(sender)

	gotID, err := handlePasswordResetCode(context.Background(), svc, marshalPayload(t, identityID))
	if err != nil {
		t.Fatalf("handlePasswordResetCode() error = %v", err)
	}
	if gotID != identityID {
		t.Fatalf("handlePasswordResetCode() = %q, want %q", gotID, identityID)
	}
	if len(sender.passwordReset) != 1 {
		t.Fatalf("SendPasswordResetEmail() called %d times, want 1", len(sender.passwordReset))
	}
}

func TestHandlePasswordChangeCode_Success(t *testing.T) {
	identityID := uuid.New().String()
	sender := &mockEmailSender{}
	svc := newNotificationService(sender)

	gotID, err := handlePasswordChangeCode(context.Background(), svc, marshalPayload(t, identityID))
	if err != nil {
		t.Fatalf("handlePasswordChangeCode() error = %v", err)
	}
	if gotID != identityID {
		t.Fatalf("handlePasswordChangeCode() = %q, want %q", gotID, identityID)
	}
	if len(sender.passwordChange) != 1 {
		t.Fatalf("SendPasswordChangeEmail() called %d times, want 1", len(sender.passwordChange))
	}
}

func TestHandleEmailChangeCode_Success(t *testing.T) {
	identityID := uuid.New().String()
	sender := &mockEmailSender{}
	svc := newNotificationService(sender)

	gotID, err := handleEmailChangeCode(context.Background(), svc, marshalPayload(t, identityID))
	if err != nil {
		t.Fatalf("handleEmailChangeCode() error = %v", err)
	}
	if gotID != identityID {
		t.Fatalf("handleEmailChangeCode() = %q, want %q", gotID, identityID)
	}
	if len(sender.emailChange) != 1 {
		t.Fatalf("SendEmailChangeEmail() called %d times, want 1", len(sender.emailChange))
	}
}

func TestHandleDeleteAccountCode_Success(t *testing.T) {
	identityID := uuid.New().String()
	sender := &mockEmailSender{}
	svc := newNotificationService(sender)

	gotID, err := handleDeleteAccountCode(context.Background(), svc, marshalPayload(t, identityID))
	if err != nil {
		t.Fatalf("handleDeleteAccountCode() error = %v", err)
	}
	if gotID != identityID {
		t.Fatalf("handleDeleteAccountCode() = %q, want %q", gotID, identityID)
	}
	if len(sender.deleteAccount) != 1 {
		t.Fatalf("SendDeleteAccountEmail() called %d times, want 1", len(sender.deleteAccount))
	}
}

func TestHandleCode_SendError(t *testing.T) {
	identityID := uuid.New().String()
	sender := &mockEmailSender{err: errors.New("smtp error")}
	svc := newNotificationService(sender)

	_, err := handlePasswordResetCode(context.Background(), svc, marshalPayload(t, identityID))
	if err == nil {
		t.Fatal("handlePasswordResetCode() error = nil, want error")
	}
}

func TestMessageHandler_Idempotent(t *testing.T) {
	eventID := uuid.New().String()
	identityID := uuid.New().String()

	pe := &mockProcessedEvents{
		existing: map[string]*records.ProcessedEvent{
			eventID: {EventID: eventID},
		},
	}
	sender := &mockEmailSender{}
	svc := newNotificationService(sender)

	log := logger.New(&config.LoggerConfig{
		Level:  slog.LevelError,
		Format: config.FormatJSON,
	})
	c := &Consumer{
		log:             log,
		svc:             svc,
		processedEvents: pe,
		handlers: map[string]eventHandler{
			"notification.email.verification_code.send": handleVerificationCode,
		},
	}

	envelope := EventEnvelope{
		ID:         eventID,
		Type:       "notification.email.verification_code.send",
		OccurredAt: time.Now(),
		Data:       marshalPayload(t, identityID),
	}
	msgData, _ := json.Marshal(envelope)

	c.messageHandler(&nats.Msg{Data: msgData})

	if len(sender.verification) != 0 {
		t.Fatalf("SendVerificationEmail() called %d times, want 0 (idempotent)", len(sender.verification))
	}
}
