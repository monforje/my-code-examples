package services

import "context"

// TagsService - сервис для работы с тегами.
type TagsService struct {
	taskRepo TaskRepository
}

// NewTagsService - конструктор TagsService.
func NewTagsService(taskRepo TaskRepository) *TagsService {
	return &TagsService{taskRepo: taskRepo}
}

// List - получает список всех тегов.
func (s *TagsService) List(ctx context.Context) (*TagsOutput, error) {
	tags, err := s.taskRepo.ListTags(ctx)
	if err != nil {
		return nil, err
	}

	return &TagsOutput{
		Items: mapTagsOutput(tags),
	}, nil
}
