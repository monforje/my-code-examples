package records

import (
	"time"

	"github.com/google/uuid"
)

type Credential struct {
	IdentityID        uuid.UUID `db:"identity_id"`
	PasswordHash      string    `db:"password_hash"`
	PasswordChangedAt time.Time `db:"password_changed_at"`
	CreatedAt         time.Time `db:"created_at"`
	UpdatedAt         time.Time `db:"updated_at"`
}
