package mailersender

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"notifications/internal/services"
	"notifications/internal/templates"
)

type mockMailer struct {
	sent []Message
	err  error
}

func (m *mockMailer) Send(_ context.Context, msg Message) error {
	if m.err != nil {
		return m.err
	}
	m.sent = append(m.sent, msg)
	return nil
}

func newRenderer(t *testing.T) *templates.Renderer {
	t.Helper()
	r, err := templates.NewRenderer()
	if err != nil {
		t.Fatalf("create renderer: %v", err)
	}
	return r
}

func newSender(t *testing.T, m Mailer) *Sender {
	t.Helper()
	return New(m, newRenderer(t), slog.Default())
}

func TestSendVerificationEmail(t *testing.T) {
	mock := &mockMailer{}
	s := newSender(t, mock)

	err := s.SendVerificationEmail(context.Background(), services.SendCodeEmailParams{
		To:   "user@example.com",
		Code: "123456",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mock.sent) != 1 {
		t.Fatalf("expected 1 email sent, got %d", len(mock.sent))
	}

	msg := mock.sent[0]
	if msg.To != "user@example.com" {
		t.Errorf("To = %q, want %q", msg.To, "user@example.com")
	}
	if msg.Subject == "" {
		t.Error("Subject is empty")
	}
	if msg.HTML == "" {
		t.Error("HTML is empty")
	}
	if msg.Text == "" {
		t.Error("Text is empty")
	}
	if !contains(msg.HTML, "123456") {
		t.Error("HTML does not contain code")
	}
}

func TestSendPasswordResetEmail(t *testing.T) {
	mock := &mockMailer{}
	s := newSender(t, mock)

	err := s.SendPasswordResetEmail(context.Background(), services.SendCodeEmailParams{
		To:   "user@example.com",
		Code: "abcdef",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mock.sent) != 1 {
		t.Fatalf("expected 1 email, got %d", len(mock.sent))
	}
	if !contains(mock.sent[0].HTML, "abcdef") {
		t.Error("HTML does not contain code")
	}
}

func TestSendPasswordChangeEmail(t *testing.T) {
	mock := &mockMailer{}
	s := newSender(t, mock)

	err := s.SendPasswordChangeEmail(context.Background(), services.SendCodeEmailParams{
		To:   "user@example.com",
		Code: "xyz789",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mock.sent) != 1 {
		t.Fatalf("expected 1 email, got %d", len(mock.sent))
	}
}

func TestSendEmailChangeEmail(t *testing.T) {
	mock := &mockMailer{}
	s := newSender(t, mock)

	err := s.SendEmailChangeEmail(context.Background(), services.SendCodeEmailParams{
		To:   "new@example.com",
		Code: "chg001",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mock.sent) != 1 {
		t.Fatalf("expected 1 email, got %d", len(mock.sent))
	}
}

func TestSendDeleteAccountEmail(t *testing.T) {
	mock := &mockMailer{}
	s := newSender(t, mock)

	err := s.SendDeleteAccountEmail(context.Background(), services.SendCodeEmailParams{
		To:   "user@example.com",
		Code: "del001",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mock.sent) != 1 {
		t.Fatalf("expected 1 email, got %d", len(mock.sent))
	}
}

func TestSendEmail_MailerError(t *testing.T) {
	mock := &mockMailer{err: errors.New("smtp connection refused")}
	s := newSender(t, mock)

	err := s.SendVerificationEmail(context.Background(), services.SendCodeEmailParams{
		To:   "user@example.com",
		Code: "123456",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !contains(err.Error(), "send email") {
		t.Errorf("error = %q, want it to contain 'send email'", err.Error())
	}
}

func TestSendEmail_WithPrivacyURL(t *testing.T) {
	mock := &mockMailer{}
	s := newSender(t, mock)

	err := s.SendVerificationEmail(context.Background(), services.SendCodeEmailParams{
		To:         "user@example.com",
		Code:       "123456",
		PrivacyURL: "https://example.com/privacy",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !contains(mock.sent[0].HTML, "https://example.com/privacy") {
		t.Error("HTML does not contain privacy URL")
	}
}

func TestSendEmail_AllTemplatesProduceContent(t *testing.T) {
	mock := &mockMailer{}
	s := newSender(t, mock)

	tests := []struct {
		name   string
		sendFn func(ctx context.Context, p services.SendCodeEmailParams) error
	}{
		{"verification", s.SendVerificationEmail},
		{"password_reset", s.SendPasswordResetEmail},
		{"password_change", s.SendPasswordChangeEmail},
		{"email_change", s.SendEmailChangeEmail},
		{"delete_account", s.SendDeleteAccountEmail},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock.sent = nil

			err := tt.sendFn(context.Background(), services.SendCodeEmailParams{
				To:   "user@example.com",
				Code: "test123",
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			msg := mock.sent[0]
			if len(msg.Subject) == 0 {
				t.Error("subject is empty")
			}
			if len(msg.HTML) < 50 {
				t.Errorf("HTML too short: %d bytes", len(msg.HTML))
			}
			if len(msg.Text) < 10 {
				t.Errorf("Text too short: %d bytes", len(msg.Text))
			}
		})
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
