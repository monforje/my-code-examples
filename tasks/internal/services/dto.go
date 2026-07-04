// Package services содержит DTO (Data Transfer Objects) для обмена данными
// между хендлерами и сервисным слоем.
package services

import (
	"time"

	"github.com/google/uuid"
)

// Create

// CreateInput - входные данные для создания задачи.
type CreateInput struct {
	TaskName            string
	Title               string
	Description         string
	SpecificationMDText string
	TaskType            string
	Level               string
	TagIDs              []uuid.UUID
	LanguageIDs         []uuid.UUID
}

// TaskOutput - выходные данные задачи.
type TaskOutput struct {
	ID                  uuid.UUID
	TaskName            string
	Title               string
	Description         string
	SpecificationMDText string
	TaskType            string
	Level               string
	Tags                []TagOutput
	Languages           []LanguageOutput
	CreatedAt           time.Time
}

// TagOutput - тег в ответе.
type TagOutput struct {
	ID   uuid.UUID
	Name string
}

// LanguageOutput - язык в ответе.
type LanguageOutput struct {
	ID   uuid.UUID
	Name string
}

// List

// ListInput - входные данные для получения списка задач.
type ListInput struct {
	Limit     int32
	Cursor    *string
	Search    *string
	Tags      []string
	Languages []string
	TaskType  *string
	Level     *string
}

// ListOutput - результат запроса списка задач.
type ListOutput struct {
	Items       []TaskListItemOutput
	HasNextPage bool
	NextCursor  *string
}

// TaskListItemOutput - элемент списка задач (без specification_md_text).
type TaskListItemOutput struct {
	ID          uuid.UUID
	TaskName    string
	Title       string
	Description string
	TaskType    string
	Level       string
	Tags        []TagOutput
	Languages   []LanguageOutput
	CreatedAt   time.Time
}

// Update

// UpdateInput - входные данные для обновления задачи.
type UpdateInput struct {
	ID                  uuid.UUID
	TaskName            *string
	Title               *string
	Description         *string
	SpecificationMDText *string
	TaskType            *string
	Level               *string
	TagIDs              *[]uuid.UUID
	LanguageIDs         *[]uuid.UUID
}

// Tags

// TagsOutput - список тегов.
type TagsOutput struct {
	Items []TagOutput
}

// Languages

// LanguagesOutput - список языков.
type LanguagesOutput struct {
	Items []LanguageOutput
}
