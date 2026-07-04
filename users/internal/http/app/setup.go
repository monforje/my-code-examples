package httpapp

import (
	httpserver "users/internal/http/gen"
	"users/internal/http/handlers"
	"users/internal/http/middleware"
	"users/internal/repository/security"
	"users/pkg/logger"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
)

type Handlers struct {
	UsersHandlers *handlers.UsersHandlers
}

func (a *App) Setup(log *logger.Logger, tokenManager *security.Manager, handlers *Handlers, serviceToken string) {
	a.e.Use(middleware.CORS(a.cfg.CORS))

	a.e.Use(echomw.RequestID())
	a.e.Use(logger.RequestLogMiddleware(log))

	if a.storage != nil {
		a.e.Static(a.storage.AvatarPublic, a.storage.AvatarDir)
	}

	protectedOps := map[string][]echo.MiddlewareFunc{
		"users.profile.me.get":             {middleware.BearerAuth(tokenManager)},
		"users.profile.me.avatar.delete":   {middleware.BearerAuth(tokenManager)},
		"users.profile.me.avatar.update":   {middleware.BearerAuth(tokenManager)},
		"users.profile.me.settings.update": {middleware.BearerAuth(tokenManager)},
		"users.git.me.get":                 {middleware.ServiceToken(serviceToken)},
	}

	httpserver.RegisterHandlersWithOptions(a.e, handlers.UsersHandlers, httpserver.RegisterHandlersOptions{
		BaseURL:              "/api/v1",
		OperationMiddlewares: protectedOps,
	})
}
