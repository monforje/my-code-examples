// Package app
package app

import (
	"context"
	"notifications/internal/config"
	httpapp "notifications/internal/http/app"
	"notifications/internal/worker/consumer"
	workerapp "notifications/internal/worker/app"
	"notifications/pkg/logger"
)

type App struct {
	diContainer *diContainer
	HTTPserver  *httpapp.App
	Worker      *workerapp.App
}

// New - конструктор приложения: собирает DI-контейнер и инициализирует worker.
func New(ctx context.Context, logger *logger.Logger, cfg *config.Config) *App {
	a := &App{
		diContainer: newDIContainer(ctx, logger, cfg),
	}

	a.initWorker()

	return a
}

// InitHTTPServer - инициализирует HTTP-сервер с маршрутов и middleware.
// Вызывается только из notifications-http.
func (a *App) InitHTTPServer() {
	a.HTTPserver = httpapp.New(a.diContainer.logger, &a.diContainer.cfg.Server)

	a.HTTPserver.Setup(a.diContainer.logger, &httpapp.Handlers{
		CiReportHandlers: a.diContainer.CIReportHandlers(),
	})
}

func (a *App) initWorker() {
	if a.diContainer.cfg.NATS.Host == "" {
		return
	}
	c := consumer.NewConsumer(
		a.diContainer.NATS().Conn(),
		a.diContainer.logger,
		a.diContainer.NotificationService(),
		a.diContainer.ProcessedEvents(),
	)
	a.Worker = workerapp.New(c)
}
