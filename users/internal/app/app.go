package app

import (
	"context"
	"users/internal/config"
	httpapp "users/internal/http/app"
	workerapp "users/internal/worker/app"
	"users/internal/worker/consumer"
	"users/pkg/logger"
)

type Mode string

const (
	ModeServer Mode = "server"
	ModeWorker Mode = "worker"
)

type App struct {
	mode        Mode
	diContainer *diContainer
	HTTPserver  *httpapp.App
	Worker      *workerapp.App
}

func New(ctx context.Context, logger *logger.Logger, cfg *config.Config, mode Mode) *App {
	a := &App{
		mode:        mode,
		diContainer: newDIContainer(ctx, logger, cfg),
	}

	a.initDeps()

	return a
}

func (a *App) initDeps() {
	switch a.mode {
	case ModeServer:
		a.initHTTPServer()
	case ModeWorker:
		a.initWorker()
	}
}

func (a *App) initHTTPServer() {
	a.HTTPserver = httpapp.New(a.diContainer.logger, &a.diContainer.cfg.Server, &a.diContainer.cfg.Storage)

	a.HTTPserver.Setup(a.diContainer.logger, a.diContainer.TokenManager(), &httpapp.Handlers{
		UsersHandlers: a.diContainer.UsersHandlers(),
	}, a.diContainer.cfg.Server.ServiceToken)
}

func (a *App) initWorker() {
	if a.diContainer.cfg.NATS.Host == "" {
		return
	}
	c := consumer.NewConsumer(
		a.diContainer.NATS().Conn(),
		a.diContainer.logger,
		a.diContainer.AuthService(),
		a.diContainer.ProcessedEvents(),
	)
	a.Worker = workerapp.New(c)
}
