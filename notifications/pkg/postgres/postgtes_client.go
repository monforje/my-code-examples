// Package postgresclient
package postgresclient

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"notifications/internal/config"
	"notifications/pkg/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultMaxOpenConns    = 25
	defaultMinConns        = 5
	defaultConnMaxLifetime = 5 * time.Minute
	defaultConnMaxIdleTime = 1 * time.Minute
	defaultPingTimeout     = 5 * time.Second
)

type Client struct {
	log *logger.Logger
	*pgxpool.Pool
}

func New(ctx context.Context, cfg config.PGConfig, log *logger.Logger) *Client {
	const op = "postgresclient.New"

	poolCfg, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		panic(fmt.Errorf("pgxpool.ParseConfig: %w", err))
	}

	poolCfg.MaxConns = defaultMaxOpenConns
	poolCfg.MinConns = defaultMinConns
	poolCfg.MaxConnLifetime = defaultConnMaxLifetime
	poolCfg.MaxConnIdleTime = defaultConnMaxIdleTime
	poolCfg.HealthCheckPeriod = defaultPingTimeout

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		panic(fmt.Errorf("pgxpool.NewWithConfig: %w", err))
	}

	if err := pool.Ping(ctx); err != nil {
		panic(fmt.Errorf("pgx ping: %w", err))
	}

	log.Info(ctx, op, "postgres client created successfully", slog.Any("dsn", cfg.DSN()))

	return &Client{log: log, Pool: pool}
}

func (c *Client) Ping(ctx context.Context) error {
	return c.Pool.Ping(ctx)
}

func (c *Client) Close() {
	c.Pool.Close()
}
