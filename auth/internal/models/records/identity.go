package records

import (
	"time"

	"github.com/google/uuid"
)

type Identity struct {
	ID            uuid.UUID  `db:"id"`
	Email         string     `db:"email"`
	EmailVerified bool       `db:"email_verified"`
	Status        string     `db:"status"`
	CreatedAt     time.Time  `db:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at"`
	DeletedAt     *time.Time `db:"deleted_at"`
}
