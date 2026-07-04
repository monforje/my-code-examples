package httpapp

import (
	httpserver "tasks/internal/http/gen"
	"tasks/internal/http/handlers"
	gittaskshandler "tasks/internal/http/handlers/git_tasks"
	reportshandler "tasks/internal/http/handlers/reports"
	"tasks/internal/http/middleware"
	"tasks/internal/repository/security"
	"tasks/pkg/logger"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
)

// Handlers - агрегатор всех HTTP-обработчиков приложения.
type Handlers struct {
	TasksHandlers   *handlers.TasksHandlers
	GitTasksHandler *gittaskshandler.GitTasksHandler
	ReportsHandler  *reportshandler.ReportsHandler
}

type serverHandlers struct {
	*handlers.TasksHandlers
	*gittaskshandler.GitTasksHandler
	*reportshandler.ReportsHandler
}

// Setup - настройка маршрутов, middleware и регистрация обработчиков.
func (a *App) Setup(log *logger.Logger, tokenManager *security.Manager, reportsServiceToken string, h *Handlers) {
	a.e.Use(middleware.CORS(a.cfg.CORS))

	a.e.Use(echomw.RequestID())
	a.e.Use(logger.RequestLogMiddleware(log))

	protectedOps := map[string][]echo.MiddlewareFunc{
		"tasks.create":         {middleware.BearerAuth(tokenManager)},
		"tasks.list":           {middleware.BearerAuth(tokenManager)},
		"tasks.get":            {middleware.BearerAuth(tokenManager)},
		"tasks.update":         {middleware.BearerAuth(tokenManager)},
		"tasks.delete":         {middleware.BearerAuth(tokenManager)},
		"tasks.tags.list":      {middleware.BearerAuth(tokenManager)},
		"tasks.languages.list": {middleware.BearerAuth(tokenManager)},
		"tasks.git.create":     {middleware.BearerAuth(tokenManager)},
		// Внутренний приём CI-отчётов от notifications — сервисный токен.
		"reports.create": {middleware.ServiceToken(reportsServiceToken)},
		// Выдача отчётов фронту — user JWT.
		"reports.list": {middleware.BearerAuth(tokenManager)},
		"reports.get":  {middleware.BearerAuth(tokenManager)},
	}

	httpserver.RegisterHandlersWithOptions(a.e, &serverHandlers{
		TasksHandlers:   h.TasksHandlers,
		GitTasksHandler: h.GitTasksHandler,
		ReportsHandler:  h.ReportsHandler,
	}, httpserver.RegisterHandlersOptions{
		BaseURL:              "/api/v1",
		OperationMiddlewares: protectedOps,
	})
}
