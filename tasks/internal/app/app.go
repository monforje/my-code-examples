// Package app реализует.bootstrap приложения и управление зависимостями.
package app

import (
	"context"

	"tasks/internal/config"
	httpapp "tasks/internal/http/app"
	"tasks/pkg/logger"
)

// App - корневая структура приложения, инкапсулирующая DI-контейнер и HTTP-сервер.
type App struct {
	diContainer *diContainer
	HTTPserver  *httpapp.App
}

// New - конструктор App.
// Инициализирует DI-контейнер и создаёт HTTP-сервер.
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

	a.HTTPserver.Setup(
		a.diContainer.logger,
		a.diContainer.TokenManager(),
		a.diContainer.cfg.Reports.ServiceToken,
		&httpapp.Handlers{
			TasksHandlers:   a.diContainer.TasksHandlers(),
			GitTasksHandler: a.diContainer.GitTasksHandler(),
			ReportsHandler:  a.diContainer.ReportsHandler(),
		},
	)
}
