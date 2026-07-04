// Package handlers предоставляет HTTP-обработчики для API задач.
//
// TasksHandlers реализует интерфейс httpserver.ServerInterface,
// обеспечивая маппинг HTTP-запросов на методы TasksService.
package handlers

import (
	"context"
	"net/http"

	httpserver "tasks/internal/http/gen"
	"tasks/internal/http/validation"
	taskservice "tasks/internal/services"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// tasksService - интерфейс сервиса задач, используемый хендлерами.
type tasksService interface {
	Create(ctx context.Context, input *taskservice.CreateInput) (*taskservice.TaskOutput, error)
	GetByID(ctx context.Context, id uuid.UUID) (*taskservice.TaskOutput, error)
	List(ctx context.Context, input *taskservice.ListInput) (*taskservice.ListOutput, error)
	Update(ctx context.Context, input *taskservice.UpdateInput) (*taskservice.TaskOutput, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// tagsService - интерфейс сервиса тегов.
type tagsService interface {
	List(ctx context.Context) (*taskservice.TagsOutput, error)
}

// languagesService - интерфейс сервиса языков.
type languagesService interface {
	List(ctx context.Context) (*taskservice.LanguagesOutput, error)
}

// TasksHandlers - обработчики HTTP-запросов для CRUD задач.
type TasksHandlers struct {
	ts  tasksService
	tgs tagsService
	ls  languagesService
}

// NewTasksHandlers - конструктор TasksHandlers.
func NewTasksHandlers(ts tasksService, tgs tagsService, ls languagesService) *TasksHandlers {
	return &TasksHandlers{ts: ts, tgs: tgs, ls: ls}
}

// mapTaskOutput - маппит TaskOutput в httpserver.Task.
func mapTaskOutput(output *taskservice.TaskOutput) httpserver.Task {
	tags := make([]httpserver.Tag, 0, len(output.Tags))
	for _, t := range output.Tags {
		tags = append(tags, httpserver.Tag{
			Id:   t.ID.String(),
			Name: t.Name,
		})
	}

	langs := make([]httpserver.Language, 0, len(output.Languages))
	for _, l := range output.Languages {
		langs = append(langs, httpserver.Language{
			Id:   l.ID.String(),
			Name: l.Name,
		})
	}

	return httpserver.Task{
		Id:                  output.ID.String(),
		TaskName:            output.TaskName,
		Title:               output.Title,
		Description:         output.Description,
		SpecificationMdText: output.SpecificationMDText,
		TaskType:            httpserver.TaskType(output.TaskType),
		Level:               httpserver.Level(output.Level),
		Tags:                tags,
		Languages:           langs,
		CreatedAt:           output.CreatedAt,
	}
}

// mapListItemOutput - маппит TaskListItemOutput в httpserver.TaskListItem.
func mapListItemOutput(item *taskservice.TaskListItemOutput) httpserver.TaskListItem {
	tags := make([]httpserver.Tag, 0, len(item.Tags))
	for _, t := range item.Tags {
		tags = append(tags, httpserver.Tag{
			Id:   t.ID.String(),
			Name: t.Name,
		})
	}

	langs := make([]httpserver.Language, 0, len(item.Languages))
	for _, l := range item.Languages {
		langs = append(langs, httpserver.Language{
			Id:   l.ID.String(),
			Name: l.Name,
		})
	}

	return httpserver.TaskListItem{
		Id:          item.ID.String(),
		TaskName:    item.TaskName,
		Title:       item.Title,
		Description: item.Description,
		TaskType:    httpserver.TaskType(item.TaskType),
		Level:       httpserver.Level(item.Level),
		Tags:        tags,
		Languages:   langs,
		CreatedAt:   item.CreatedAt,
	}
}

// TasksCreate - обработчик создания задачи
func (h *TasksHandlers) TasksCreate(ctx echo.Context) error {
	var req httpserver.CreateTaskRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.ErrBadRequest
	}

	title, spec, description, taskType, level, tagIDs, langIDs, err := validation.ValidateCreateTaskRequest(
		req.Title, req.SpecificationMdText, req.Description,
		string(req.TaskType), string(req.Level),
		req.TagIds, req.LanguageIds,
	)
	if err != nil {
		return validationError(ctx, err)
	}

	input := &taskservice.CreateInput{
		TaskName:            req.TaskName,
		Title:               title,
		Description:         description,
		SpecificationMDText: spec,
		TaskType:            taskType,
		Level:               level,
		TagIDs:              tagIDs,
		LanguageIDs:         langIDs,
	}

	output, err := h.ts.Create(ctx.Request().Context(), input)
	if err != nil {
		return serviceError(ctx, err)
	}

	return ctx.JSON(http.StatusCreated, mapTaskOutput(output))
}

// TasksGet - обработчик получения задачи по ID
func (h *TasksHandlers) TasksGet(ctx echo.Context, taskId string) error {
	id, err := validation.ValidateTaskID(taskId)
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, httpserver.VALIDATIONERROR, err)
	}

	output, err := h.ts.GetByID(ctx.Request().Context(), id)
	if err != nil {
		return serviceError(ctx, err)
	}

	return ctx.JSON(http.StatusOK, mapTaskOutput(output))
}

// TasksList - обработчик получения списка задач с курсорной пагинацией
func (h *TasksHandlers) TasksList(ctx echo.Context, params httpserver.TasksListParams) error {
	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}

	input := &taskservice.ListInput{
		Limit:     limit,
		Cursor:    params.Cursor,
		Search:    params.Search,
		Tags:      derefStringSlice(params.Tags),
		Languages: derefStringSlice(params.Languages),
		TaskType:  derefTaskType(params.TaskType),
		Level:     derefLevel(params.Level),
	}

	output, err := h.ts.List(ctx.Request().Context(), input)
	if err != nil {
		return serviceError(ctx, err)
	}

	items := make([]httpserver.TaskListItem, 0, len(output.Items))
	for _, item := range output.Items {
		items = append(items, mapListItemOutput(&item))
	}

	return ctx.JSON(http.StatusOK, httpserver.TaskListResponse{
		Items: items,
		PageInfo: httpserver.PageInfo{
			HasNextPage: output.HasNextPage,
			NextCursor:  output.NextCursor,
		},
	})
}

// TasksUpdate - обработчик обновления задачи
func (h *TasksHandlers) TasksUpdate(ctx echo.Context, taskId string) error {
	id, err := validation.ValidateTaskID(taskId)
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, httpserver.VALIDATIONERROR, err)
	}

	var req httpserver.UpdateTaskRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.ErrBadRequest
	}

	title, spec, description, taskType, level, tagIDs, langIDs, err := validation.ValidateUpdateTaskRequest(
		derefString(req.Title), derefString(req.SpecificationMdText), derefString(req.Description),
		derefString((*string)(req.TaskType)), derefString((*string)(req.Level)),
		derefStringSlice(req.TagIds), derefStringSlice(req.LanguageIds),
	)
	if err != nil {
		return validationError(ctx, err)
	}

	input := &taskservice.UpdateInput{
		ID:                  id,
		TaskName:            req.TaskName,
		Title:               title,
		Description:         description,
		SpecificationMDText: spec,
		TaskType:            taskType,
		Level:               level,
		TagIDs:              tagIDs,
		LanguageIDs:         langIDs,
	}

	output, err := h.ts.Update(ctx.Request().Context(), input)
	if err != nil {
		return serviceError(ctx, err)
	}

	return ctx.JSON(http.StatusOK, mapTaskOutput(output))
}

// TasksDelete - обработчик удаления задачи
func (h *TasksHandlers) TasksDelete(ctx echo.Context, taskId string) error {
	id, err := validation.ValidateTaskID(taskId)
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, httpserver.VALIDATIONERROR, err)
	}

	if err := h.ts.Delete(ctx.Request().Context(), id); err != nil {
		return serviceError(ctx, err)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// TasksTagsList - обработчик получения списка тегов
func (h *TasksHandlers) TasksTagsList(ctx echo.Context) error {
	output, err := h.tgs.List(ctx.Request().Context())
	if err != nil {
		return serviceError(ctx, err)
	}

	tags := make([]httpserver.Tag, 0, len(output.Items))
	for _, t := range output.Items {
		tags = append(tags, httpserver.Tag{
			Id:   t.ID.String(),
			Name: t.Name,
		})
	}

	return ctx.JSON(http.StatusOK, tags)
}

// TasksLanguagesList - обработчик получения списка языков
func (h *TasksHandlers) TasksLanguagesList(ctx echo.Context) error {
	output, err := h.ls.List(ctx.Request().Context())
	if err != nil {
		return serviceError(ctx, err)
	}

	langs := make([]httpserver.Language, 0, len(output.Items))
	for _, l := range output.Items {
		langs = append(langs, httpserver.Language{
			Id:   l.ID.String(),
			Name: l.Name,
		})
	}

	return ctx.JSON(http.StatusOK, langs)
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func derefStringSlice(s *[]string) []string {
	if s == nil {
		return nil
	}
	return *s
}

func derefTaskType(t *httpserver.TaskType) *string {
	if t == nil {
		return nil
	}
	s := string(*t)
	return &s
}

func derefLevel(l *httpserver.Level) *string {
	if l == nil {
		return nil
	}
	s := string(*l)
	return &s
}
