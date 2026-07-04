package services

import (
	"context"

	"github.com/google/uuid"

	"tasks/internal/models/records"
)

// Create - создает новую задачу.
/*
	1. Генерация нового UUID для задачи.
	2. Формирование текущего времени.
	3. Вставка записи в БД.
	4. Установка тегов и языков.
	5. Возвращение данных созданной задачи.
*/
func (s *TasksService) Create(ctx context.Context, input *CreateInput) (*TaskOutput, error) {
	if input.Title == "" {
		return nil, ErrTitleEmpty
	}

	now := nowFunc()
	task := &records.Task{
		ID:                  uuid.New(),
		TaskName:            input.TaskName,
		Title:               input.Title,
		Description:         input.Description,
		SpecificationMDText: input.SpecificationMDText,
		TaskType:            input.TaskType,
		Level:               input.Level,
		CreatedAt:           now,
	}

	if err := s.taskRepo.Create(ctx, task); err != nil {
		return nil, err
	}

	if len(input.TagIDs) > 0 {
		if err := s.taskRepo.SetTags(ctx, task.ID, input.TagIDs); err != nil {
			return nil, err
		}
	}

	if len(input.LanguageIDs) > 0 {
		if err := s.taskRepo.SetLanguages(ctx, task.ID, input.LanguageIDs); err != nil {
			return nil, err
		}
	}

	return s.buildTaskOutput(ctx, task)
}
