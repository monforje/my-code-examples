// Package services
package services

import (
	"context"
)

type EmailSender interface {
	SendVerificationEmail(ctx context.Context, params SendCodeEmailParams) error
	SendPasswordResetEmail(ctx context.Context, params SendCodeEmailParams) error
	SendPasswordChangeEmail(ctx context.Context, params SendCodeEmailParams) error
	SendEmailChangeEmail(ctx context.Context, params SendCodeEmailParams) error
	SendDeleteAccountEmail(ctx context.Context, params SendCodeEmailParams) error
}

type SendCodeEmailParams struct {
	To             string
	Code           string
	PrivacyURL     string
	CompanyAddress string
}

type NotificationService struct {
	sender EmailSender
}

func NewNotificationService(sender EmailSender) *NotificationService {
	return &NotificationService{sender: sender}
}
