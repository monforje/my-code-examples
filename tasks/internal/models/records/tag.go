package records

import "github.com/google/uuid"

// Tag - тег задачи.
type Tag struct {
	ID   uuid.UUID `db:"id"`
	Name string    `db:"name"`
}
