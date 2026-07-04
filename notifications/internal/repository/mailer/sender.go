// Package mailersender provides the repository layer for sending emails.
package mailersender

import (
	"context"
	"fmt"
	"notifications/internal/services"
	"notifications/internal/templates"
	"notifications/pkg/mailer"

	"log/slog"
)

type Message struct {
	To      string
	Subject string
	HTML    string
	Text    string
}

type Mailer interface {
	Send(ctx context.Context, msg Message) error
}

type mailerAdapter struct {
	client *mailer.Client
}

func NewMailerAdapter(c *mailer.Client) Mailer {
	return &mailerAdapter{client: c}
}

func (a *mailerAdapter) Send(ctx context.Context, msg Message) error {
	return a.client.Send(ctx, mailer.Message{
		To:      msg.To,
		Subject: msg.Subject,
		HTML:    msg.HTML,
		Text:    msg.Text,
	})
}

type Sender struct {
	mailer   Mailer
	renderer *templates.Renderer
	log      *slog.Logger
}

func New(m Mailer, r *templates.Renderer, log *slog.Logger) *Sender {
	return &Sender{
		mailer:   m,
		renderer: r,
		log:      log,
	}
}

func (s *Sender) SendVerificationEmail(ctx context.Context, params services.SendCodeEmailParams) error {
	return s.sendCodeEmail(ctx, templates.CodeVerification, params)
}

func (s *Sender) SendPasswordResetEmail(ctx context.Context, params services.SendCodeEmailParams) error {
	return s.sendCodeEmail(ctx, templates.CodePasswordReset, params)
}

func (s *Sender) SendPasswordChangeEmail(ctx context.Context, params services.SendCodeEmailParams) error {
	return s.sendCodeEmail(ctx, templates.CodePasswordChange, params)
}

func (s *Sender) SendEmailChangeEmail(ctx context.Context, params services.SendCodeEmailParams) error {
	return s.sendCodeEmail(ctx, templates.CodeEmailChange, params)
}

func (s *Sender) SendDeleteAccountEmail(ctx context.Context, params services.SendCodeEmailParams) error {
	return s.sendCodeEmail(ctx, templates.CodeDeleteAccount, params)
}

func (s *Sender) sendCodeEmail(ctx context.Context, tpl templates.CodeTemplate, params services.SendCodeEmailParams) error {
	rendered, err := s.renderer.RenderCodeEmail(tpl, templates.CodeEmailData{
		Email:          params.To,
		Code:           params.Code,
		PrivacyURL:     params.PrivacyURL,
		CompanyAddress: params.CompanyAddress,
	})
	if err != nil {
		return fmt.Errorf("render email template %q: %w", tpl, err)
	}

	msg := Message{
		To:      params.To,
		Subject: rendered.Subject,
		HTML:    rendered.HTML,
		Text:    rendered.Text,
	}

	if err := s.mailer.Send(ctx, msg); err != nil {
		return fmt.Errorf("send email to %s: %w", params.To, err)
	}

	s.log.InfoContext(ctx, "email sent",
		slog.String("to", params.To),
		slog.String("template", string(tpl)),
	)

	return nil
}
