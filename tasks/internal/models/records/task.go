package records

import (
	"time"

	"github.com/google/uuid"
)

// Task - доменная модель задачи для передачи между слоями (postgres → services).
type Task struct {
	ID                  uuid.UUID `db:"id"`
	TaskName            string    `db:"task_name"`
	Title               string    `db:"title"`
	Description         string    `db:"description"`
	SpecificationMDText string    `db:"specification_md_text"`
	TaskType            string    `db:"task_type"`
	Level               string    `db:"level"`
	CreatedAt           time.Time `db:"created_at"`
}

// TaskListItem - элемент списка задач (без specification_md_text).
type TaskListItem struct {
	ID          uuid.UUID `db:"id"`
	TaskName    string    `db:"task_name"`
	Title       string    `db:"title"`
	Description string    `db:"description"`
	TaskType    string    `db:"task_type"`
	Level       string    `db:"level"`
	CreatedAt   time.Time `db:"created_at"`
}
