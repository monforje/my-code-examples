package services

import (
	"context"

	"github.com/google/uuid"
)

// GetByID - получает задачу по UUID.
/*
	1. Получение задачи из БД по ID.
	2. Если задача не найдена - возврат ErrTaskNotFound.
	3. Загрузка связей (tags, languages).
	4. Маппинг в TaskOutput.
*/
func (s *TasksService) GetByID(ctx context.Context, id uuid.UUID) (*TaskOutput, error) {
	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrTaskNotFound
	}

	return s.buildTaskOutput(ctx, task)
}
