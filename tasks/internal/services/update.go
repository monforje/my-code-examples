package services

import "context"

// Update - обновляет заголовок и/или спецификацию задачи.
/*
	1. Получение текущей задачи из БД по ID.
	2. Если задача не найдена - возврат ErrTaskNotFound.
	3. Применение изменений: если поле передано - обновляем, иначе оставляем как есть.
	4. Обновление записи в БД.
	5. Обновление связей (tags, languages) если переданы.
*/
func (s *TasksService) Update(ctx context.Context, input *UpdateInput) (*TaskOutput, error) {
	task, err := s.taskRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, ErrTaskNotFound
	}

	if input.TaskName != nil {
		task.TaskName = *input.TaskName
	}
	if input.Title != nil {
		task.Title = *input.Title
	}
	if input.Description != nil {
		task.Description = *input.Description
	}
	if input.SpecificationMDText != nil {
		task.SpecificationMDText = *input.SpecificationMDText
	}
	if input.TaskType != nil {
		task.TaskType = *input.TaskType
	}
	if input.Level != nil {
		task.Level = *input.Level
	}

	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, err
	}

	if input.TagIDs != nil {
		if err := s.taskRepo.SetTags(ctx, task.ID, *input.TagIDs); err != nil {
			return nil, err
		}
	}

	if input.LanguageIDs != nil {
		if err := s.taskRepo.SetLanguages(ctx, task.ID, *input.LanguageIDs); err != nil {
			return nil, err
		}
	}

	return s.buildTaskOutput(ctx, task)
}
