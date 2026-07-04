// Package app
package app

import (
	"auth/internal/config"
	httpapp "auth/internal/http/app"
	"auth/pkg/logger"
	"context"
)

type App struct {
	diContainer *diContainer
	HTTPserver  *httpapp.App
}

func New(ctx context.Context, logger *logger.Logger, cfg *config.Config) *App {
	a := &App{
		diContainer: newDIContainer(ctx, logger, cfg),
	}

	a.initDeps()

	return a
}

func (a *App) initDeps() {
	inits := []func(){
		a.initHTTPServer,
	}

	for _, fn := range inits {
		fn()
	}
}

func (a *App) initHTTPServer() {
	a.HTTPserver = httpapp.New(a.diContainer.logger, &a.diContainer.cfg.Server)

	a.HTTPserver.Setup(a.diContainer.logger, a.diContainer.TokenManager(), &httpapp.Handlers{
		AuthHandlers:    a.diContainer.AuthHandlers(),
		AuthCliHandlers: a.diContainer.AuthCliHandlers(),
	})
}
