package gittaskshandler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"go.uber.org/mock/gomock"

	"tasks/internal/authctx"
	httpserver "tasks/internal/http/gen"
	gittaskshandler "tasks/internal/http/handlers/git_tasks"
	"tasks/internal/http/handlers/git_tasks/mocks"
	gittasksservice "tasks/internal/services/git_tasks"
)

func TestTasksGitCreate_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockGitTasksService(ctrl)
	h := gittaskshandler.NewGitTasksHandler(svc)

	svc.EXPECT().CreateGitTask(gomock.Any(), &gittasksservice.GitTaskCreateInput{TaskName: "pizza-api"}).Return(&gittasksservice.GitTaskCreateOutput{
		TaskName: "pizza-api",
		Repo:     "alice/golden-pizza-api",
		CloneURL: "http://alice:token@gitea.local/alice/golden-pizza-api.git",
	}, nil)

	body, _ := json.Marshal(httpserver.GitTaskRequest{TaskName: "pizza-api"})
	req := httptest.NewRequest(http.MethodPost, "/tasks/pizza-api/git", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := echo.New().NewContext(req, rec)

	if err := h.TasksGitCreate(ctx, "pizza-api"); err != nil {
		t.Fatalf("TasksGitCreate() error = %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}

	var resp httpserver.GitTaskResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.CloneUrl == "" {
		t.Fatal("clone_url is empty")
	}
}

func TestTasksGitCreate_TaskNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockGitTasksService(ctrl)
	h := gittaskshandler.NewGitTasksHandler(svc)
	svc.EXPECT().CreateGitTask(gomock.Any(), gomock.Any()).Return(nil, gittasksservice.ErrTaskNotFound)

	req := httptest.NewRequest(http.MethodPost, "/tasks/missing/git", nil)
	rec := httptest.NewRecorder()
	ctx := echo.New().NewContext(req, rec)

	if err := h.TasksGitCreate(ctx, "missing"); err != nil {
		t.Fatalf("TasksGitCreate() error = %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestTasksGitCreate_MissingAuthContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockGitTasksService(ctrl)
	h := gittaskshandler.NewGitTasksHandler(svc)
	svc.EXPECT().CreateGitTask(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, _ *gittasksservice.GitTaskCreateInput) (*gittasksservice.GitTaskCreateOutput, error) {
		return nil, authctx.ErrAuthContextMissing
	})

	req := httptest.NewRequest(http.MethodPost, "/tasks/pizza-api/git", nil)
	rec := httptest.NewRecorder()
	ctx := echo.New().NewContext(req, rec)

	if err := h.TasksGitCreate(ctx, "pizza-api"); err != nil {
		t.Fatalf("TasksGitCreate() error = %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestTasksGitCreate_InternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockGitTasksService(ctrl)
	h := gittaskshandler.NewGitTasksHandler(svc)
	svc.EXPECT().CreateGitTask(gomock.Any(), gomock.Any()).Return(nil, errors.New("boom"))

	req := httptest.NewRequest(http.MethodPost, "/tasks/pizza-api/git", nil)
	rec := httptest.NewRecorder()
	ctx := echo.New().NewContext(req, rec)

	if err := h.TasksGitCreate(ctx, "pizza-api"); err != nil {
		t.Fatalf("TasksGitCreate() error = %v", err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}
