package services

import (
	"context"

	"tasks/internal/models/records"
)

// List - получает страницу задач с курсорной пагинацией и фильтрами.
/*
	1. Установка лимита по умолчанию (20), если не задан.
	2. Ограничение лимита максимумом 100.
	3. Формирование фильтров.
	4. Вызов репозитория для получения списка.
	5. Загрузка связей (tags, languages) для каждого элемента.
	6. Формирование cursor для следующей страницы.
*/
func (s *TasksService) List(ctx context.Context, input *ListInput) (*ListOutput, error) {
	limit := input.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	filters := ListFilters{
		Search:    input.Search,
		Tags:      input.Tags,
		Languages: input.Languages,
		TaskType:  input.TaskType,
		Level:     input.Level,
	}

	items, hasNextPage, err := s.taskRepo.List(ctx, limit, input.Cursor, filters)
	if err != nil {
		return nil, err
	}

	result := make([]TaskListItemOutput, 0, len(items))
	for _, item := range items {
		tags, err := s.taskRepo.GetTagsByTaskID(ctx, item.ID)
		if err != nil {
			return nil, err
		}

		langs, err := s.taskRepo.GetLanguagesByTaskID(ctx, item.ID)
		if err != nil {
			return nil, err
		}

		result = append(result, TaskListItemOutput{
			ID:          item.ID,
			TaskName:    item.TaskName,
			Title:       item.Title,
			Description: item.Description,
			TaskType:    item.TaskType,
			Level:       item.Level,
			Tags:        mapTagsOutput(tags),
			Languages:   mapLanguagesOutput(langs),
			CreatedAt:   item.CreatedAt,
		})
	}

	var nextCursor *string
	if hasNextPage && len(result) > 0 {
		cursor := result[len(result)-1].ID.String()
		nextCursor = &cursor
	}

	return &ListOutput{
		Items:       result,
		HasNextPage: hasNextPage,
		NextCursor:  nextCursor,
	}, nil
}

func mapTagsOutput(tags []records.Tag) []TagOutput {
	result := make([]TagOutput, 0, len(tags))
	for _, t := range tags {
		result = append(result, TagOutput{ID: t.ID, Name: t.Name})
	}
	return result
}

func mapLanguagesOutput(langs []records.Language) []LanguageOutput {
	result := make([]LanguageOutput, 0, len(langs))
	for _, l := range langs {
		result = append(result, LanguageOutput{ID: l.ID, Name: l.Name})
	}
	return result
}
