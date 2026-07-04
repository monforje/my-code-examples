package natsrepo

import (
	"context"
	"encoding/json"
	"fmt"

	"auth/internal/events"

	"github.com/nats-io/nats.go"
)

type natsClient interface {
	PublishMsg(msg *nats.Msg) error
}

type Producer struct {
	conn natsClient
}

func NewProducer(conn natsClient) *Producer {
	return &Producer{conn: conn}
}

func (p *Producer) Publish(ctx context.Context, subject string, key []byte, event events.Event) error {
	value, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal nats event: %w", err)
	}

	msg := &nats.Msg{
		Subject: subject,
		Data:    value,
		Header:  nats.Header{"X-Key": []string{string(key)}},
	}

	if err := p.conn.PublishMsg(msg); err != nil {
		return fmt.Errorf("publish nats event %s: %w", event.Type, err)
	}

	return nil
}

func (p *Producer) PublishIdentityCreated(ctx context.Context, payload events.IdentityCreatedPayload) error {
	return p.Publish(ctx, string(events.EventIdentityCreated), []byte(payload.IdentityID), events.NewEvent(events.EventIdentityCreated, payload))
}

func (p *Producer) PublishIdentityUpdated(ctx context.Context, payload events.IdentityUpdatedPayload) error {
	return p.Publish(ctx, string(events.EventIdentityUpdated), []byte(payload.IdentityID), events.NewEvent(events.EventIdentityUpdated, payload))
}

func (p *Producer) PublishIdentityDeleted(ctx context.Context, payload events.IdentityDeletedPayload) error {
	return p.Publish(ctx, string(events.EventIdentityDeleted), []byte(payload.IdentityID), events.NewEvent(events.EventIdentityDeleted, payload))
}

func (p *Producer) PublishIdentityLogin(ctx context.Context, payload events.IdentityLoginPayload) error {
	return p.Publish(ctx, string(events.EventIdentityLogin), []byte(payload.IdentityID), events.NewEvent(events.EventIdentityLogin, payload))
}

func (p *Producer) PublishIdentityLogout(ctx context.Context, payload events.IdentityLogoutPayload) error {
	return p.Publish(ctx, string(events.EventIdentityLogout), []byte(payload.IdentityID), events.NewEvent(events.EventIdentityLogout, payload))
}

func (p *Producer) PublishVerificationCodeSend(ctx context.Context, payload events.VerificationCodeSendPayload) error {
	return p.Publish(ctx, string(events.EventVerificationCodeSend), []byte(payload.Email), events.NewEvent(events.EventVerificationCodeSend, payload))
}

func (p *Producer) PublishPasswordResetCodeSend(ctx context.Context, payload events.VerificationCodeSendPayload) error {
	return p.Publish(ctx, string(events.EventPasswordResetCodeSend), []byte(payload.Email), events.NewEvent(events.EventPasswordResetCodeSend, payload))
}

func (p *Producer) PublishPasswordChangeCodeSend(ctx context.Context, payload events.VerificationCodeSendPayload) error {
	return p.Publish(ctx, string(events.EventPasswordChangeCodeSend), []byte(payload.Email), events.NewEvent(events.EventPasswordChangeCodeSend, payload))
}

func (p *Producer) PublishEmailChangeCodeSend(ctx context.Context, payload events.VerificationCodeSendPayload) error {
	return p.Publish(ctx, string(events.EventEmailChangeCodeSend), []byte(payload.Email), events.NewEvent(events.EventEmailChangeCodeSend, payload))
}

func (p *Producer) PublishAccountDeleteCodeSend(ctx context.Context, payload events.VerificationCodeSendPayload) error {
	return p.Publish(ctx, string(events.EventAccountDeleteCodeSend), []byte(payload.Email), events.NewEvent(events.EventAccountDeleteCodeSend, payload))
}


