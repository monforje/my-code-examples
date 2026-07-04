// Package redisclient
package redisclient

import (
	"context"
	"fmt"

	"notifications/internal/config"
	"notifications/pkg/logger"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	log *logger.Logger
	*redis.Client
}

func New(ctx context.Context, cfg config.RedisConfig, log *logger.Logger) *Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		panic(fmt.Errorf("redis ping: %w", err))
	}

	log.Info(ctx, "redis.New", "redis connected successfully", "addr", cfg.Addr())

	return &Client{Client: rdb}
}

func (c *Client) Ping(ctx context.Context) error {
	return c.Client.Ping(ctx).Err()
}
