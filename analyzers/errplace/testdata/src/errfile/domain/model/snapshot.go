// Positive + boundary: snapshot.go is NOT a file for errors.
// Declaring an error variable here violates GID-169, while ordinary
// (non-error) variables and custom error types are acceptable.
package model

import "github.com/pkg/errors"

// --- Positive: an error var in a wrong file ---

var ErrSnapshotConflict = errors.New("snapshot conflict") // want `GID-169: error "ErrSnapshotConflict" is declared in snapshot\.go\. Fix: keep layer errors in error\.go`

// --- Boundary: a type implementing error via a pointer ---

// ValidationError implements error ON THE POINTER.
type ValidationError struct{ Field string }

func (e *ValidationError) Error() string { return e.Field }

// errValidation has type *ValidationError → implements error → a violation.
var errValidation = &ValidationError{} // want `GID-169: error "errValidation" is declared in snapshot\.go\. Fix: keep layer errors in error\.go`

// errValidationValue has type ValidationError (value): the Error method is
// declared on the pointer, the value does NOT implement error → not a violation.
var errValidationValue = ValidationError{}

// --- Boundary: non-error package-level variables — out of the rule's scope ---

var DefaultLimit = 100

var snapshotName = "snapshot"

type Snapshot struct{ ID string }
