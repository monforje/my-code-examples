package httpapp

import (
	"strings"

	httpserver "notifications/internal/http/gen"
	"notifications/internal/http/handlers"
	"notifications/internal/http/middleware"
	"notifications/pkg/logger"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
)

// Handlers - агрегатор всех HTTP-обработчиков приложения.
type Handlers struct {
	CiReportHandlers *handlers.CiReportHandlers
}

// Setup - настройка маршрутов, middleware и регистрация обработчиков.
func (a *App) Setup(log *logger.Logger, h *Handlers) {
	// task_runner шлёт POST на baseURL + "/". Strip trailing slash без редиректа.
	a.e.Pre(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			p := c.Request().URL.Path
			if len(p) > 1 && p[len(p)-1] == '/' {
				c.Request().URL.Path = strings.TrimRight(p, "/")
			}
			return next(c)
		}
	})

	a.e.Use(middleware.CORS(a.cfg.CORS))
	a.e.Use(echomw.RequestID())
	a.e.Use(logger.RequestLogMiddleware(log))

	httpserver.RegisterHandlersWithOptions(a.e, h.CiReportHandlers, httpserver.RegisterHandlersOptions{
		BaseURL: "/api/v1/notifications",
	})
}
