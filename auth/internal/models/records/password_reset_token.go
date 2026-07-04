package records

import (
	"time"

	"github.com/google/uuid"
)

type PasswordResetToken struct {
	ID         uuid.UUID  `db:"id"`
	IdentityID uuid.UUID  `db:"identity_id"`
	TokenHash  string     `db:"token_hash"`
	ExpiresAt  time.Time  `db:"expires_at"`
	ConsumedAt *time.Time `db:"consumed_at"`
	CreatedAt  time.Time  `db:"created_at"`
}
