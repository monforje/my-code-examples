// Package consumer
package consumer

import (
	"context"
	"encoding/json"
	"errors"

	"log/slog"
	"time"

	"notifications/internal/models/records"
	postgresrepo "notifications/internal/repository/postgres"
	service "notifications/internal/services"
	"notifications/pkg/logger"

	"github.com/google/uuid"

	"github.com/nats-io/nats.go"
)

type EventEnvelope struct {
	ID         string          `json:"id"`
	Type       string          `json:"type"`
	OccurredAt time.Time       `json:"occurred_at"`
	Data       json.RawMessage `json:"data"`
}

type eventHandler func(ctx context.Context, svc *service.NotificationService, data []byte) (string, error)

type processedEventsRepository interface {
	Create(ctx context.Context, event *records.ProcessedEvent) error
	GetByEventID(ctx context.Context, eventID string) (*records.ProcessedEvent, error)
}

type Consumer struct {
	conn            *nats.Conn
	log             *logger.Logger
	svc             *service.NotificationService
	processedEvents processedEventsRepository
	subs            []*nats.Subscription
	handlers        map[string]eventHandler
	done            chan struct{}
}

func NewConsumer(
	conn *nats.Conn,
	log *logger.Logger,
	svc *service.NotificationService,
	processedEvents processedEventsRepository,
) *Consumer {
	return &Consumer{
		conn:            conn,
		log:             log,
		svc:             svc,
		processedEvents: processedEvents,
		done:            make(chan struct{}),
		handlers: map[string]eventHandler{
			"notification.email.verification_code.send":    handleVerificationCode,
			"notification.email.password_reset_code.send":  handlePasswordResetCode,
			"notification.email.password_change_code.send": handlePasswordChangeCode,
			"notification.email.email_change_code.send":    handleEmailChangeCode,
			"notification.email.account_delete_code.send":  handleDeleteAccountCode,
		},
	}
}

// Run — подписывается на NATS subjects и начинает обработку сообщений.
func (c *Consumer) Run() error {
	const op = "Consumer.Run"

	for subj := range c.handlers {
		sub, err := c.conn.Subscribe(subj, c.messageHandler)
		if err != nil {
			return errors.New(op + ": subscribe " + subj + ": " + err.Error())
		}
		c.subs = append(c.subs, sub)
		c.log.Info(context.Background(), op, "subscribed", slog.String("subject", subj))
	}

	<-c.done
	return nil
}

// messageHandler — обработчик NATS сообщения.
// Последовательность: проверка идемпотентности → вызов handler → запись обработанного события.
func (c *Consumer) messageHandler(msg *nats.Msg) {
	const op = "Consumer.messageHandler"

	var envelope EventEnvelope
	if err := json.Unmarshal(msg.Data, &envelope); err != nil {
		c.log.Error(context.Background(), op, "unmarshal event", "error", err)
		return
	}

	handler, ok := c.handlers[envelope.Type]
	if !ok {
		c.log.Warn(context.Background(), op, "unknown event type", slog.String("type", envelope.Type))
		return
	}

	ctx := context.Background()

	// 1. Проверка идемпотентности.
	_, err := c.processedEvents.GetByEventID(ctx, envelope.ID)
	if err == nil {
		return
	}
	if !errors.Is(err, postgresrepo.ErrProcessedEventNotFound) {
		c.log.Error(ctx, op, "check idempotency", "error", err)
		return
	}

	// 2. Вызов обработчика.
	identityID, err := handler(ctx, c.svc, envelope.Data)
	if err != nil {
		c.log.Error(ctx, op, "handle event", slog.String("type", envelope.Type), "error", err)
		return
	}

	c.log.Info(ctx, op, "event handled",
		slog.String("type", envelope.Type),
		slog.String("event_id", envelope.ID),
		slog.String("identity_id", identityID),
	)

	// 3. Запись обработанного события.
	aggID, err := uuid.Parse(identityID)
	if err != nil {
		c.log.Error(ctx, op, "parse aggregate id", "error", err)
		return
	}

	if err := c.processedEvents.Create(ctx, &records.ProcessedEvent{
		EventID:     envelope.ID,
		EventType:   envelope.Type,
		AggregateID: aggID,
		ProcessedAt: time.Now().UTC(),
	}); err != nil {
		c.log.Error(ctx, op, "record processed event", "error", err)
	}
}

// Stop — отписывается от всех subjects.
func (c *Consumer) Stop() {
	const op = "Consumer.Stop"

	for _, sub := range c.subs {
		if err := sub.Unsubscribe(); err != nil {
			c.log.Error(context.Background(), op, "unsubscribe", "error", err)
		}
	}
	close(c.done)
	c.log.Info(context.Background(), op, "consumer stopped")
}
