package records

import (
	"net"
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID               uuid.UUID  `db:"id"`
	IdentityID       uuid.UUID  `db:"identity_id"`
	RefreshTokenHash string     `db:"refresh_token_hash"`
	UserAgent        string     `db:"user_agent"`
	IPAddress        *net.IP    `db:"ip_address"`
	ExpiresAt        time.Time  `db:"expires_at"`
	RevokedAt        *time.Time `db:"revoked_at"`
	CreatedAt        time.Time  `db:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at"`
}
