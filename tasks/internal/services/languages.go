package services

import "context"

// LanguagesService - сервис для работы с языками.
type LanguagesService struct {
	taskRepo TaskRepository
}

// NewLanguagesService - конструктор LanguagesService.
func NewLanguagesService(taskRepo TaskRepository) *LanguagesService {
	return &LanguagesService{taskRepo: taskRepo}
}

// List - получает список всех языков.
func (s *LanguagesService) List(ctx context.Context) (*LanguagesOutput, error) {
	langs, err := s.taskRepo.ListLanguages(ctx)
	if err != nil {
		return nil, err
	}

	return &LanguagesOutput{
		Items: mapLanguagesOutput(langs),
	}, nil
}
