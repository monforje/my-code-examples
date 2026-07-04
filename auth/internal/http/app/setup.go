package httpapp

import (
	httpserver "auth/internal/http/gen"
	"auth/internal/http/handlers"
	authclihandlers "auth/internal/http/handlers/auth_cli"
	"auth/internal/http/middleware"
	"auth/internal/repository/security"
	"auth/pkg/logger"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
)

type Handlers struct {
	AuthHandlers    *handlers.AuthHandlers
	AuthCliHandlers *authclihandlers.AuthCliHandlers
}

// combinedHandler implements ServerInterface by delegating to both handler structs.
type combinedHandler struct {
	*handlers.AuthHandlers
	cli *authclihandlers.AuthCliHandlers
}

func (c *combinedHandler) AuthDeviceStart(ctx echo.Context) error {
	return c.cli.AuthDeviceStart(ctx)
}

func (c *combinedHandler) AuthDeviceConfirm(ctx echo.Context) error {
	return c.cli.AuthDeviceConfirm(ctx)
}

func (c *combinedHandler) AuthDeviceToken(ctx echo.Context) error {
	return c.cli.AuthDeviceToken(ctx)
}

func (c *combinedHandler) AuthCliRefresh(ctx echo.Context) error {
	return c.cli.AuthCliRefresh(ctx)
}

func (a *App) Setup(log *logger.Logger, tokenManager *security.Manager, handlers *Handlers) {
	a.e.Use(middleware.CORS(a.cfg.CORS))

	a.e.Use(echomw.RequestID())
	a.e.Use(logger.RequestLogMiddleware(log))

	protectedOps := map[string][]echo.MiddlewareFunc{
		"auth.me.get":                      {middleware.BearerAuth(tokenManager)},
		"auth.me.delete":                   {middleware.BearerAuth(tokenManager)},
		"auth.me.delete.verify":            {middleware.BearerAuth(tokenManager)},
		"auth.me.delete.code.resend":       {middleware.BearerAuth(tokenManager)},
		"auth.me.email.change":             {middleware.BearerAuth(tokenManager)},
		"auth.me.email.change.verify":      {middleware.BearerAuth(tokenManager)},
		"auth.me.email.change.confirm":     {middleware.BearerAuth(tokenManager)},
		"auth.me.email.change.complete":    {middleware.BearerAuth(tokenManager)},
		"auth.me.email.change.code.resend": {middleware.BearerAuth(tokenManager)},
		"auth.password.change":             {middleware.BearerAuth(tokenManager)},
		"auth.password.change.verify":      {middleware.BearerAuth(tokenManager)},
		"auth.password.change.complete":    {middleware.BearerAuth(tokenManager)},
		"auth.password.change.code.resend": {middleware.BearerAuth(tokenManager)},
		"auth.logout":                      {middleware.BearerAuth(tokenManager)},
		"auth.device.confirm":              {middleware.BearerAuth(tokenManager)},
	}

	combined := &combinedHandler{
		AuthHandlers: handlers.AuthHandlers,
		cli:          handlers.AuthCliHandlers,
	}

	httpserver.RegisterHandlersWithOptions(a.e, combined, httpserver.RegisterHandlersOptions{
		BaseURL:              "/api/v1",
		OperationMiddlewares: protectedOps,
	})
}
