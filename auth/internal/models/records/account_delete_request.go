// Package records
package records

import (
	"time"

	"github.com/google/uuid"
)

type AccountDeleteRequest struct {
	ID         uuid.UUID `db:"id"`
	IdentityID uuid.UUID `db:"identity_id"`
	Status     string    `db:"status"`
	ExpiresAt  time.Time `db:"expires_at"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}
