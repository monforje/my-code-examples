package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/mock/gomock"

	httpserver "tasks/internal/http/gen"
	"tasks/internal/http/handlers"
	mocks "tasks/internal/http/handlers/mocks"
	"tasks/internal/services"
)

func newEcho() *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	return e
}

func TestTasksCreate_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockTasksService(ctrl)
	tgs := mocks.NewMockTagsService(ctrl)
	ls := mocks.NewMockLanguagesService(ctrl)
	h := handlers.NewTasksHandlers(svc, tgs, ls)

	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	svc.EXPECT().Create(gomock.Any(), &services.CreateInput{
		Title:               "New Task",
		Description:         "Description",
		SpecificationMDText: "# Spec",
		TaskType:            "backend",
		Level:               "middle",
	}).Return(&services.TaskOutput{
		ID:                  uuid.New(),
		Title:               "New Task",
		Description:         "Description",
		SpecificationMDText: "# Spec",
		TaskType:            "backend",
		Level:               "middle",
		CreatedAt:           now,
	}, nil)

	body, _ := json.Marshal(httpserver.CreateTaskRequest{
		Title:               "New Task",
		Description:         "Description",
		SpecificationMdText: "# Spec",
		TaskType:            "backend",
		Level:               "middle",
		TagIds:              []string{},
		LanguageIds:         []string{},
	})
	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e := newEcho()
	ctx := e.NewContext(req, rec)

	if err := h.TasksCreate(ctx); err != nil {
		t.Fatalf("TasksCreate() error = %v", err)
	}

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}

	var resp httpserver.Task
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Title != "New Task" {
		t.Errorf("Title = %q, want %q", resp.Title, "New Task")
	}
}

func TestTasksCreate_EmptyTitle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockTasksService(ctrl)
	tgs := mocks.NewMockTagsService(ctrl)
	ls := mocks.NewMockLanguagesService(ctrl)
	h := handlers.NewTasksHandlers(svc, tgs, ls)

	body, _ := json.Marshal(httpserver.CreateTaskRequest{Title: ""})
	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e := newEcho()
	ctx := e.NewContext(req, rec)

	if err := h.TasksCreate(ctx); err != nil {
		t.Fatalf("TasksCreate() error = %v", err)
	}

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnprocessableEntity)
	}
}

func TestTasksGet_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockTasksService(ctrl)
	tgs := mocks.NewMockTagsService(ctrl)
	ls := mocks.NewMockLanguagesService(ctrl)
	h := handlers.NewTasksHandlers(svc, tgs, ls)

	id := uuid.New()
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	svc.EXPECT().GetByID(gomock.Any(), id).Return(&services.TaskOutput{
		ID:        id,
		Title:     "found",
		TaskType:  "backend",
		Level:     "middle",
		CreatedAt: now,
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	e := newEcho()
	ctx := e.NewContext(req, rec)

	if err := h.TasksGet(ctx, id.String()); err != nil {
		t.Fatalf("TasksGet() error = %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestTasksGet_InvalidUUID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockTasksService(ctrl)
	tgs := mocks.NewMockTagsService(ctrl)
	ls := mocks.NewMockLanguagesService(ctrl)
	h := handlers.NewTasksHandlers(svc, tgs, ls)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	e := newEcho()
	ctx := e.NewContext(req, rec)

	if err := h.TasksGet(ctx, "not-a-uuid"); err != nil {
		t.Fatalf("TasksGet() error = %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestTasksGet_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockTasksService(ctrl)
	tgs := mocks.NewMockTagsService(ctrl)
	ls := mocks.NewMockLanguagesService(ctrl)
	h := handlers.NewTasksHandlers(svc, tgs, ls)

	id := uuid.New()
	svc.EXPECT().GetByID(gomock.Any(), id).Return(nil, services.ErrTaskNotFound)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	e := newEcho()
	ctx := e.NewContext(req, rec)

	if err := h.TasksGet(ctx, id.String()); err != nil {
		t.Fatalf("TasksGet() error = %v", err)
	}

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestTasksList_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockTasksService(ctrl)
	tgs := mocks.NewMockTagsService(ctrl)
	ls := mocks.NewMockLanguagesService(ctrl)
	h := handlers.NewTasksHandlers(svc, tgs, ls)

	svc.EXPECT().List(gomock.Any(), &services.ListInput{Limit: 20}).Return(&services.ListOutput{
		Items: []services.TaskListItemOutput{
			{ID: uuid.New(), Title: "Task 1", TaskType: "backend", Level: "middle", CreatedAt: time.Now()},
			{ID: uuid.New(), Title: "Task 2", TaskType: "frontend", Level: "senior", CreatedAt: time.Now()},
		},
		HasNextPage: false,
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/tasks?limit=20", nil)
	rec := httptest.NewRecorder()

	e := newEcho()
	ctx := e.NewContext(req, rec)

	if err := h.TasksList(ctx, httpserver.TasksListParams{Limit: 20}); err != nil {
		t.Fatalf("TasksList() error = %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp httpserver.TaskListResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Items) != 2 {
		t.Errorf("items count = %d, want 2", len(resp.Items))
	}
}

func TestTasksUpdate_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockTasksService(ctrl)
	tgs := mocks.NewMockTagsService(ctrl)
	ls := mocks.NewMockLanguagesService(ctrl)
	h := handlers.NewTasksHandlers(svc, tgs, ls)

	id := uuid.New()
	newTitle := "Updated"
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	svc.EXPECT().Update(gomock.Any(), &services.UpdateInput{
		ID:    id,
		Title: &newTitle,
	}).Return(&services.TaskOutput{
		ID:        id,
		Title:     "Updated",
		TaskType:  "backend",
		Level:     "middle",
		CreatedAt: now,
	}, nil)

	body, _ := json.Marshal(httpserver.UpdateTaskRequest{
		Title: &newTitle,
	})
	req := httptest.NewRequest(http.MethodPatch, "/tasks", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e := newEcho()
	ctx := e.NewContext(req, rec)

	if err := h.TasksUpdate(ctx, id.String()); err != nil {
		t.Fatalf("TasksUpdate() error = %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestTasksDelete_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockTasksService(ctrl)
	tgs := mocks.NewMockTagsService(ctrl)
	ls := mocks.NewMockLanguagesService(ctrl)
	h := handlers.NewTasksHandlers(svc, tgs, ls)

	id := uuid.New()
	svc.EXPECT().Delete(gomock.Any(), id).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()

	e := newEcho()
	ctx := e.NewContext(req, rec)

	if err := h.TasksDelete(ctx, id.String()); err != nil {
		t.Fatalf("TasksDelete() error = %v", err)
	}

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestTasksDelete_InvalidUUID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockTasksService(ctrl)
	tgs := mocks.NewMockTagsService(ctrl)
	ls := mocks.NewMockLanguagesService(ctrl)
	h := handlers.NewTasksHandlers(svc, tgs, ls)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()

	e := newEcho()
	ctx := e.NewContext(req, rec)

	if err := h.TasksDelete(ctx, "not-a-uuid"); err != nil {
		t.Fatalf("TasksDelete() error = %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestTasksDelete_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockTasksService(ctrl)
	tgs := mocks.NewMockTagsService(ctrl)
	ls := mocks.NewMockLanguagesService(ctrl)
	h := handlers.NewTasksHandlers(svc, tgs, ls)

	id := uuid.New()
	svc.EXPECT().Delete(gomock.Any(), id).Return(services.ErrTaskNotFound)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()

	e := newEcho()
	ctx := e.NewContext(req, rec)

	if err := h.TasksDelete(ctx, id.String()); err != nil {
		t.Fatalf("TasksDelete() error = %v", err)
	}

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
