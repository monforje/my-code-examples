package records

import (
	"time"

	"github.com/google/uuid"
)

type VerificationCode struct {
	ID            uuid.UUID  `db:"id"`
	IdentityID    *uuid.UUID `db:"identity_id"`
	Email         *string    `db:"email"`
	Purpose       string     `db:"purpose"`
	CodeHash      string     `db:"code_hash"`
	AttemptsCount int32      `db:"attempts_count"`
	MaxAttempts   int32      `db:"max_attempts"`
	ExpiresAt     time.Time  `db:"expires_at"`
	ConsumedAt    *time.Time `db:"consumed_at"`
	CreatedAt     time.Time  `db:"created_at"`
}
