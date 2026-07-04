// Package workerapp
package workerapp

import (
	"context"

	"notifications/internal/worker/consumer"
)

type App struct {
	consumer *consumer.Consumer
}

func New(c *consumer.Consumer) *App {
	return &App{
		consumer: c,
	}
}

func (a *App) Run() error {
	return a.consumer.Run()
}

func (a *App) Shutdown(_ context.Context) {
	a.consumer.Stop()
}
