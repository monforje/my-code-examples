package records

import (
	"time"

	"github.com/google/uuid"
)

// PulledTask - запись о пуленном репозитории задачи.
type PulledTask struct {
	ID         uuid.UUID `db:"id"`
	IdentityID uuid.UUID `db:"identity_id"`
	TaskID     uuid.UUID `db:"task_id"`
	Repo       string    `db:"repo"`
	CloneURL   string    `db:"clone_url"`
	CreatedAt  time.Time `db:"created_at"`
}
