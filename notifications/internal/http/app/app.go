package httpapp

import (
	"context"
	"net"

	"notifications/internal/config"
	"notifications/pkg/logger"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
)

// App - HTTP-сервер на базе Echo.
type App struct {
	e   *echo.Echo
	cfg *config.ServerConfig
	log *logger.Logger
}

// New - конструктор App.
func New(log *logger.Logger, cfg *config.ServerConfig) *App {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Use(echomw.RecoverWithConfig(echomw.RecoverConfig{
		DisablePrintStack: true,
		DisableStackAll:   true,
	}))

	return &App{
		e:   e,
		cfg: cfg,
		log: log,
	}
}

// Run - запуск HTTP-сервера на порту из конфигурации.
func (a *App) Run() error {
	return a.e.Start(":" + a.cfg.Port)
}

// RunOnListener - запуск HTTP-сервера на конкретном listener (для e2e тестов).
func (a *App) RunOnListener(listener net.Listener) error {
	return a.e.Server.Serve(listener)
}

// Shutdown - корректное завершение HTTP-сервера.
func (a *App) Shutdown(ctx context.Context) error {
	return a.e.Shutdown(ctx)
}
