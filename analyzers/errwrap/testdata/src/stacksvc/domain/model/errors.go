// Static errors of the model package for the GID-248 eval.
package model

import "github.com/pkg/errors"

// ErrNotFound — a static error: it carries no stack of its own.
var ErrNotFound = errors.New("not found")

// BigError — a named error type; a literal of it carries no stack either.
type BigError struct{ Code int }

func (e *BigError) Error() string { return "big error" }
