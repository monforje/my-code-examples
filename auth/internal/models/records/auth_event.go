package records

import (
	"net"
	"time"

	"github.com/google/uuid"
)

type AuthEvent struct {
	ID         uuid.UUID  `db:"id"`
	IdentityID *uuid.UUID `db:"identity_id"`
	EventType  string     `db:"event_type"`
	IPAddress  *net.IP    `db:"ip_address"`
	UserAgent  string     `db:"user_agent"`
	Metadata   []byte     `db:"metadata"`
	CreatedAt  time.Time  `db:"created_at"`
}
