package services

import (
	"context"
	"errors"
	"testing"
)

type mockEmailSender struct {
	verification   []SendCodeEmailParams
	passwordReset  []SendCodeEmailParams
	passwordChange []SendCodeEmailParams
	emailChange    []SendCodeEmailParams
	deleteAccount  []SendCodeEmailParams
	err            error
}

func (m *mockEmailSender) SendVerificationEmail(_ context.Context, params SendCodeEmailParams) error {
	if m.err != nil {
		return m.err
	}
	m.verification = append(m.verification, params)
	return nil
}

func (m *mockEmailSender) SendPasswordResetEmail(_ context.Context, params SendCodeEmailParams) error {
	if m.err != nil {
		return m.err
	}
	m.passwordReset = append(m.passwordReset, params)
	return nil
}

func (m *mockEmailSender) SendPasswordChangeEmail(_ context.Context, params SendCodeEmailParams) error {
	if m.err != nil {
		return m.err
	}
	m.passwordChange = append(m.passwordChange, params)
	return nil
}

func (m *mockEmailSender) SendEmailChangeEmail(_ context.Context, params SendCodeEmailParams) error {
	if m.err != nil {
		return m.err
	}
	m.emailChange = append(m.emailChange, params)
	return nil
}

func (m *mockEmailSender) SendDeleteAccountEmail(_ context.Context, params SendCodeEmailParams) error {
	if m.err != nil {
		return m.err
	}
	m.deleteAccount = append(m.deleteAccount, params)
	return nil
}

func TestNotificationService_SendVerificationCode(t *testing.T) {
	sender := &mockEmailSender{}
	svc := NewNotificationService(sender)

	err := svc.SendVerificationCode(context.Background(), SendCodeEmailParams{To: "user@example.com", Code: "123456"})
	if err != nil {
		t.Fatalf("SendVerificationCode() error = %v", err)
	}
	if len(sender.verification) != 1 {
		t.Fatalf("SendVerificationEmail() called %d times, want 1", len(sender.verification))
	}
}

func TestNotificationService_SendPasswordResetCode(t *testing.T) {
	sender := &mockEmailSender{}
	svc := NewNotificationService(sender)

	err := svc.SendPasswordResetCode(context.Background(), SendCodeEmailParams{To: "user@example.com", Code: "123456"})
	if err != nil {
		t.Fatalf("SendPasswordResetCode() error = %v", err)
	}
	if len(sender.passwordReset) != 1 {
		t.Fatalf("SendPasswordResetEmail() called %d times, want 1", len(sender.passwordReset))
	}
}

func TestNotificationService_SendPasswordChangeCode(t *testing.T) {
	sender := &mockEmailSender{}
	svc := NewNotificationService(sender)

	err := svc.SendPasswordChangeCode(context.Background(), SendCodeEmailParams{To: "user@example.com", Code: "123456"})
	if err != nil {
		t.Fatalf("SendPasswordChangeCode() error = %v", err)
	}
	if len(sender.passwordChange) != 1 {
		t.Fatalf("SendPasswordChangeEmail() called %d times, want 1", len(sender.passwordChange))
	}
}

func TestNotificationService_SendEmailChangeCode(t *testing.T) {
	sender := &mockEmailSender{}
	svc := NewNotificationService(sender)

	err := svc.SendEmailChangeCode(context.Background(), SendCodeEmailParams{To: "user@example.com", Code: "123456"})
	if err != nil {
		t.Fatalf("SendEmailChangeCode() error = %v", err)
	}
	if len(sender.emailChange) != 1 {
		t.Fatalf("SendEmailChangeEmail() called %d times, want 1", len(sender.emailChange))
	}
}

func TestNotificationService_SendDeleteAccountCode(t *testing.T) {
	sender := &mockEmailSender{}
	svc := NewNotificationService(sender)

	err := svc.SendDeleteAccountCode(context.Background(), SendCodeEmailParams{To: "user@example.com", Code: "123456"})
	if err != nil {
		t.Fatalf("SendDeleteAccountCode() error = %v", err)
	}
	if len(sender.deleteAccount) != 1 {
		t.Fatalf("SendDeleteAccountEmail() called %d times, want 1", len(sender.deleteAccount))
	}
}

func TestNotificationService_SendError(t *testing.T) {
	wantErr := errors.New("smtp error")
	svc := NewNotificationService(&mockEmailSender{err: wantErr})

	err := svc.SendVerificationCode(context.Background(), SendCodeEmailParams{To: "user@example.com", Code: "123456"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("SendVerificationCode() error = %v, want %v", err, wantErr)
	}
}
