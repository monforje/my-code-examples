package records

import (
	"time"

	"github.com/google/uuid"
)

type DeviceAuthorizationCode struct {
	ID             uuid.UUID  `db:"id"`
	DeviceCodeHash string     `db:"device_code_hash"`
	UserCode       string     `db:"user_code"`
	IdentityID     *uuid.UUID `db:"identity_id"`
	Status         string     `db:"status"`
	ExpiresAt      time.Time  `db:"expires_at"`
	Interval       int        `db:"interval"`
	CreatedAt      time.Time  `db:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at"`
	ConfirmedAt    *time.Time `db:"confirmed_at"`
	LastPolledAt   *time.Time `db:"last_polled_at"`
}
