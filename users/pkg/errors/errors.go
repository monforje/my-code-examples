// Package apperrors
package apperrors

type Error struct {
	Op  string
	Err error
}

func (e *Error) Error() string {
	return e.Op + ": " + e.Err.Error()
}

func (e *Error) Unwrap() error {
	return e.Err
}

func New(op string, err error) error {
	return newError(op, err)
}

func newError(op string, err error) *Error {
	return &Error{
		Op:  op,
		Err: err,
	}
}
