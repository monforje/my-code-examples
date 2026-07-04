package workerapp

import (
	"context"

	"users/internal/worker/consumer"
)

type App struct {
	consumer *consumer.Consumer
}

// New создаёт worker app.
func New(c *consumer.Consumer) *App {
	return &App{
		consumer: c,
	}
}

// Run — запуск consumer'а (подписка на NATS subjects, блокировка goroutine).
/*
	1. Вызвать a.consumer.Run().
	2. Вернуть ошибку, если подписка не удалась.
*/
func (a *App) Run() error {
	return a.consumer.Run()
}

// Shutdown — graceful shutdown: отписка от всех subjects.
func (a *App) Shutdown(_ context.Context) {
	a.consumer.Stop()
}
