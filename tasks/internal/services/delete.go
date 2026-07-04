package services

import (
	"context"

	"github.com/google/uuid"
)

// Delete - удаляет задачу по UUID.
/*
	1. Вызов репозитория для удаления записи из БД.
	2. Если задача не найдена - возврат ErrTaskNotFound.
*/
func (s *TasksService) Delete(ctx context.Context, id uuid.UUID) error {
	err := s.taskRepo.Delete(ctx, id)
	if err != nil {
		return ErrTaskNotFound
	}
	return nil
}
