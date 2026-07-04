// Package validation
package validation

import "errors"

var (
	ErrTitleEmpty      = errors.New("title is empty")
	ErrTitleTooLong    = errors.New("title must be at most 255 characters")
	ErrTaskIDInvalid   = errors.New("invalid task ID format")
	ErrInvalidTaskType = errors.New("invalid task type, must be 'backend' or 'frontend'")
	ErrInvalidLevel    = errors.New("invalid level, must be 'junior', 'middle', or 'senior'")
	ErrInvalidUUID     = errors.New("invalid UUID format")
)
