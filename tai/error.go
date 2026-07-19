package main

import "fmt"

// UserError carries a message written for the end user, while still wrapping the
// underlying technical cause so it remains available for machine-readable output
// and for errors.Is/errors.As inspection.
type UserError struct {
	Message string // Friendly, user-facing text. May be capitalised and formatted.
	Cause   error  // The underlying technical error, or nil.
}

func (e *UserError) Error() string { return e.Message }

func (e *UserError) Unwrap() error { return e.Cause }

// userErrorf builds a UserError with a formatted, user-facing message that wraps
// the given cause (which may be nil).
func userErrorf(cause error, format string, a ...any) *UserError {
	return &UserError{Message: fmt.Sprintf(format, a...), Cause: cause}
}
