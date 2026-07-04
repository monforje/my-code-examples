package records

import (
	"time"

	"github.com/google/uuid"
)

type EmailChangeRequest struct {
	ID         uuid.UUID  `db:"id"`
	IdentityID uuid.UUID  `db:"identity_id"`
	NewEmail   string     `db:"new_email"`
	Status     string     `db:"status"`
	TokenHash  *string    `db:"token_hash"`
	ExpiresAt  time.Time  `db:"expires_at"`
	ConsumedAt *time.Time `db:"consumed_at"`
	CreatedAt  time.Time  `db:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at"`
}
