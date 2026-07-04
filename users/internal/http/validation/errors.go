// Package validation
package validation

import "errors"

var (
	ErrDisplayNameEmpty   = errors.New("display name is empty")
	ErrDisplayNameTooLong = errors.New("display name must be at most 50 characters")
	ErrBioTooLong         = errors.New("bio must be at most 500 characters")
	ErrAvatarEmpty        = errors.New("avatar file is empty")
	ErrAvatarTooLarge     = errors.New("avatar file must be at most 5 MB")
	ErrAvatarInvalidType  = errors.New("avatar must be JPEG, PNG or WebP")
)
