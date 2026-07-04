package records

import "github.com/google/uuid"

// Language - язык задачи.
type Language struct {
	ID   uuid.UUID `db:"id"`
	Name string    `db:"name"`
}
