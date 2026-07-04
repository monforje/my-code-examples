package validation

import "errors"

var (
	ErrEmailEmpty         = errors.New("email is empty")
	ErrEmailInvalidFormat = errors.New("invalid email format")
	ErrEmailDomainNoMX    = errors.New("email domain does not accept mail")

	ErrPasswordEmpty    = errors.New("password is empty")
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
	ErrPasswordTooLong  = errors.New("password must be at most 72 characters")
	ErrPasswordNoLetter = errors.New("password must contain at least one letter")
	ErrPasswordNoDigit  = errors.New("password must contain at least one digit")

	ErrCodeEmpty  = errors.New("code is empty")
	ErrTokenEmpty = errors.New("token is empty")
)
