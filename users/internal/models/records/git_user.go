package records

import (
	"time"

	"github.com/google/uuid"
)

type GitUser struct {
	ID        uuid.UUID `db:"id"`
	ProfileID uuid.UUID `db:"profile_id"`
	GitToken  string    `db:"git_token"`
	GitURL    string    `db:"git_url"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
