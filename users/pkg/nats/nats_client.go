package nats

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"users/internal/config"
	"users/pkg/logger"

	"github.com/nats-io/nats.go"
)

const (
	defaultConnectTimeout = 5 * time.Second
	defaultMaxReconnects  = -1
	defaultReconnectWait  = 2 * time.Second
)

// Client - обёртка над NATS-соединением и JetStream с graceful shutdown.
type Client struct {
	conn *nats.Conn
	log  *logger.Logger
}

// NewClient подключается к NATS.
func New(ctx context.Context, cfg config.NATSConfig, log *logger.Logger) *Client {
	const op = "nats.NewClient"

	opts := []nats.Option{
		nats.Timeout(defaultConnectTimeout),
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(defaultMaxReconnects),
		nats.ReconnectWait(defaultReconnectWait),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			log.Warn(ctx, "nats.new", "nats disconnected", "error", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Info(ctx, "nats.new", "nats reconnected", "url", nc.ConnectedUrl())
		}),
		nats.ErrorHandler(func(_ *nats.Conn, _ *nats.Subscription, err error) {
			log.Error(ctx, "nats.new", "nats async error", "error", err)
		}),
	}

	url := cfg.URL()
	conn, err := nats.Connect(url, opts...)
	if err != nil {
		panic(fmt.Errorf("failed to connect to nats: %w", err))
	}

	if !conn.IsConnected() {
		panic(fmt.Errorf("failed to connect to nats: not connected"))
	}

	log.Info(ctx, op, "nats connected", "url", slog.String("url", url))

	return &Client{conn: conn, log: log}
}

// Close выполняет graceful shutdown: Drain + Close.
func (c *Client) Close(ctx context.Context) {
	const op = "nats.Client.Close"

	if c.conn == nil {
		return
	}
	c.log.Info(ctx, op, "closing nats connection")
	if err := c.conn.Drain(); err != nil {
		c.log.Error(ctx, op, "nats drain error", "error", err)
	}
	c.conn.Close()
}

// Conn возвращает raw NATS connection.
func (c *Client) Conn() *nats.Conn {
	return c.conn
}

func (c *Client) IsConnected() bool {
	return c.conn != nil && c.conn.IsConnected()
}
