// Package gittaskshandler
package gittaskshandler

import (
	"context"
	"errors"
	"log"
	"net/http"

	"tasks/internal/authctx"
	httpserver "tasks/internal/http/gen"
	gittasksservice "tasks/internal/services/git_tasks"

	"github.com/labstack/echo/v4"
)

type gitTasksService interface {
	CreateGitTask(
		ctx context.Context,
		input *gittasksservice.GitTaskCreateInput,
	) (*gittasksservice.GitTaskCreateOutput, error)
}

type GitTasksHandler struct {
	s gitTasksService
}

func NewGitTasksHandler(s gitTasksService) *GitTasksHandler {
	return &GitTasksHandler{
		s: s,
	}
}

// TasksGitCreate - Создаёт/подготавливает task repo и возвращает clone_url.
/*
	1. Проверить Bearer token middleware.
	2. Достать user_id из JWT из контекста.
	3. Вызвать сервис TasksGitCreate()
	4. Вернуть GitTaskResponse.
*/
func (h *GitTasksHandler) TasksGitCreate(ctx echo.Context, taskName string) error {
	var req httpserver.GitTaskRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.ErrBadRequest
	}
	if req.TaskName != "" {
		taskName = req.TaskName
	}

	output, err := h.s.CreateGitTask(ctx.Request().Context(), &gittasksservice.GitTaskCreateInput{TaskName: taskName})
	if err != nil {
		return gitTaskServiceError(ctx, err)
	}
	log.Printf("Repo: %s, CloneURL: %s", output.Repo, output.CloneURL)
	return ctx.JSON(http.StatusCreated, httpserver.GitTaskResponse{
		TaskName: output.TaskName,
		Repo:     output.Repo,
		CloneUrl: output.CloneURL,
	})
}

func gitTaskServiceError(ctx echo.Context, err error) error {
	status := http.StatusInternalServerError
	code := httpserver.INTERNALERROR

	switch {
	case errors.Is(err, gittasksservice.ErrTaskNotFound):
		status = http.StatusNotFound
		code = httpserver.NOTFOUND
	case errors.Is(err, authctx.ErrAuthContextMissing):
		status = http.StatusUnauthorized
		code = httpserver.MISSINGAUTHTOKEN
	}

	return ctx.JSON(status, httpserver.ErrorResponse{Code: code, Message: err.Error()})
}
