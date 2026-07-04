// Package services предоставляет бизнес-логику для работы с задачами.
//
// TasksService отвечает за CRUD-операции над задачами:
//   - Create: создание новой задачи
//   - GetByID: получение задачи по идентификатору
//   - List: получение списка задач с курсорной пагинацией
//   - Update: обновление заголовка и спецификации задачи
//   - Delete: удаление задачи
package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"tasks/internal/models/records"
)

var (
	// ErrTaskNotFound - задача с указанным ID не найдена
	ErrTaskNotFound = errors.New("task not found")

	// ErrTitleEmpty - заголовок задачи не может быть пустым
	ErrTitleEmpty = errors.New("title must not be empty")
)

// ListFilters - фильтры для списка задач.
type ListFilters struct {
	Search    *string
	Tags      []string
	Languages []string
	TaskType  *string
	Level     *string
}

// TaskRepository - интерфейс доступа к хранилищу задач.
// Определяет контракт для чтения и записи задач в БД.
type TaskRepository interface {
	Create(ctx context.Context, task *records.Task) error
	GetByID(ctx context.Context, id uuid.UUID) (*records.Task, error)
	Update(ctx context.Context, task *records.Task) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit int32, cursor *string, filters ListFilters) ([]records.TaskListItem, bool, error)
	GetTagsByTaskID(ctx context.Context, taskID uuid.UUID) ([]records.Tag, error)
	SetTags(ctx context.Context, taskID uuid.UUID, tagIDs []uuid.UUID) error
	GetLanguagesByTaskID(ctx context.Context, taskID uuid.UUID) ([]records.Language, error)
	SetLanguages(ctx context.Context, taskID uuid.UUID, languageIDs []uuid.UUID) error
	GetTagsByIDs(ctx context.Context, ids []uuid.UUID) ([]records.Tag, error)
	GetLanguagesByIDs(ctx context.Context, ids []uuid.UUID) ([]records.Language, error)
	ListTags(ctx context.Context) ([]records.Tag, error)
	ListLanguages(ctx context.Context) ([]records.Language, error)
}

// TasksService - сервис для работы с задачами.
// Инкапсулирует бизнес-логику и delegation в репозиторий.
type TasksService struct {
	taskRepo TaskRepository
}

// NewTasksService - конструктор TasksService.
func NewTasksService(taskRepo TaskRepository) *TasksService {
	return &TasksService{taskRepo: taskRepo}
}

// buildTaskOutput - маппит records.Task в TaskOutput с загрузкой связей.
func (s *TasksService) buildTaskOutput(ctx context.Context, task *records.Task) (*TaskOutput, error) {
	tags, err := s.taskRepo.GetTagsByTaskID(ctx, task.ID)
	if err != nil {
		return nil, err
	}

	langs, err := s.taskRepo.GetLanguagesByTaskID(ctx, task.ID)
	if err != nil {
		return nil, err
	}

	return &TaskOutput{
		ID:                  task.ID,
		TaskName:            task.TaskName,
		Title:               task.Title,
		Description:         task.Description,
		SpecificationMDText: task.SpecificationMDText,
		TaskType:            task.TaskType,
		Level:               task.Level,
		Tags:                mapTagsOutput(tags),
		Languages:           mapLanguagesOutput(langs),
		CreatedAt:           task.CreatedAt,
	}, nil
}

// nowFunc - функция получения текущего времени. Переопределяется в тестах.
var nowFunc = func() time.Time { return time.Now() }

// SetNowFunc - установка функции времени для тестов.
func SetNowFunc(fn func() time.Time) {
	if fn == nil {
		nowFunc = func() time.Time { return time.Now() }
		return
	}
	nowFunc = fn
}
