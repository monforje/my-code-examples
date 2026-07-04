package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"tasks/internal/app"
	"tasks/internal/app/closer"
	"tasks/internal/config"
	"tasks/pkg/logger"
	"time"
)

const (
	closerShutdownTimeout = 15 * time.Second
)

func main() {
	const op = "main"

	appEnv, err := config.ParseAppEnv(os.Getenv("APP_ENV"))
	if err != nil {
		panic(err)
	}

	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	if err := cfg.Validate(); err != nil {
		panic(err)
	}

	log := logger.New(config.NewLoggerConfig(appEnv))

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	application := app.New(ctx, log, cfg)

	log.Info(ctx, op, "starting http", slog.String("env", string(appEnv)), slog.String("port", cfg.Server.Port))

	serverErr := make(chan error, 1)

	go func() {
		serverErr <- application.HTTPserver.Run()
	}()

	log.Info(ctx, op, "server started successfully")

	select {
	case <-ctx.Done():
		log.Info(ctx, op, "shutdown signal received")
	case err := <-serverErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error(ctx, op, "server stopped with error", "error", err)
		}
	}

	stop()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), closerShutdownTimeout)
	defer shutdownCancel()

	if err := application.HTTPserver.Shutdown(shutdownCtx); err != nil {
		log.Error(shutdownCtx, op, "http shutdown error", "error", err)
	}

	log.Info(context.Background(), op, "server stopped, closing resources")

	if err := closer.CloseAll(shutdownCtx); err != nil {
		log.Error(shutdownCtx, op, "error during shutdown", "error", err)
	}

	log.Info(context.Background(), op, "application stopped gracefully")
}
