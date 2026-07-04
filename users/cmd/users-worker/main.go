package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
	"users/internal/app"
	"users/internal/app/closer"
	"users/internal/config"
	"users/pkg/logger"
)

const closerShutdownTimeout = 15 * time.Second

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

	application := app.New(ctx, log, cfg, app.ModeWorker)

	log.Info(ctx, op, "starting worker", slog.String("env", string(appEnv)))

	workerErr := make(chan error, 1)

	go func() {
		workerErr <- application.Worker.Run()
	}()

	log.Info(ctx, op, "worker started successfully")

	select {
	case <-ctx.Done():
		log.Info(ctx, op, "shutdown signal received")
	case err := <-workerErr:
		if err != nil {
			log.Error(ctx, op, "worker stopped with error", "error", err)
		}
	}

	stop()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), closerShutdownTimeout)
	defer shutdownCancel()

	application.Worker.Shutdown(shutdownCtx)

	log.Info(context.Background(), op, "worker stopped, closing resources")

	if err := closer.CloseAll(shutdownCtx); err != nil {
		log.Error(shutdownCtx, op, "error during shutdown", "error", err)
	}

	log.Info(context.Background(), op, "worker stopped gracefully")
}
