package httpapp

import (
	"auth/internal/config"
	"auth/pkg/logger"
	"context"
	"net"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
)

type App struct {
	e   *echo.Echo
	cfg *config.ServerConfig
	log *logger.Logger
}

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

func (a *App) Run() error {
	return a.e.Start(":" + a.cfg.Port)
}

func (a *App) RunOnListener(listener net.Listener) error {
	return a.e.Server.Serve(listener)
}

func (a *App) Shutdown(ctx context.Context) error {
	return a.e.Shutdown(ctx)
}
