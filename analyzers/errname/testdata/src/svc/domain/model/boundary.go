package model

import "errors"

// Boundary: a name from the list but not an error — the rule does not apply.
var ErrNotFoundMessage = "not found"

// Boundary: a blank identifier is skipped.
var _ = errors.New("blank")

// Boundary: a local variable in a function — not package-level.
func localErr() error {
	errNotFound := errors.New("not found")
	var ErrInvalid = errors.New("invalid")
	_ = ErrInvalid
	return errNotFound
}

// Boundary: a custom type implementing error via a pointer.
type validationError struct{}

func (e *validationError) Error() string { return "validation" }

// The pointer implements error → a name from the list is reported.
var ErrInvalid = &validationError{} // want `GID-234: generic error name "ErrInvalid" in domain model\. Fix: bind it to the entity: ErrSnapshotInvalid`

// A value (not a pointer) does not implement error → not reported.
var ErrFailed = validationError{}
