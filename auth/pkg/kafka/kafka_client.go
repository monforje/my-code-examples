// Package kafkaclient
package kafkaclient

import (
	"context"
	"fmt"
	"log/slog"

	"auth/internal/config"
	"auth/pkg/logger"

	"github.com/twmb/franz-go/pkg/kgo"
)

type Client struct {
	log *logger.Logger
	*kgo.Client
}

func New(ctx context.Context, cfg config.KafkaConfig, log *logger.Logger) *Client {
	const op = "kafkaclient.New"

	client, err := kgo.NewClient(
		kgo.SeedBrokers(cfg.Brokers...),
	)
	if err != nil {
		panic(fmt.Errorf("kgo.NewClient: %w", err))
	}

	if err := client.Ping(ctx); err != nil {
		panic(fmt.Errorf("kgo.Ping: %w", err))
	}

	log.Info(ctx, op, "kafka client created successfully", slog.Any("brokers", cfg.Brokers))

	return &Client{Client: client}
}

func (c *Client) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 1)
	defer cancel()

	return c.Client.Ping(ctx)
}
