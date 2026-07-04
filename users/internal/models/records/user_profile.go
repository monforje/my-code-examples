// Package records
package records

import (
	"time"

	"github.com/google/uuid"
)

type UserProfile struct {
	ID              uuid.UUID  `db:"id"`
	IdentityID      uuid.UUID  `db:"identity_id"`
	Email           string     `db:"email"`
	DisplayName     string     `db:"display_name"`
	BIO             string     `db:"bio"`
	AvatarURL       string     `db:"avatar_url"`
	AvatarObjectKey string     `db:"avatar_object_key"`
	Status          string     `db:"status"`
	EmailVerified   bool       `db:"email_verified"`
	CreatedAt       time.Time  `db:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at"`
	DeletedAt       *time.Time `db:"deleted_at"`
}
