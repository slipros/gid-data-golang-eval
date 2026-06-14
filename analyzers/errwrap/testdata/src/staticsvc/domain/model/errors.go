// Static model errors: a package-level var and a named error type.
// var declarations are NOT touched by GID-177 (they are not returns).
package model

import "github.com/pkg/errors"

var ErrSnapshotNotFound = errors.New("snapshot not found")

// BigError — a named error type.
type BigError struct {
	Code int
}

func (e BigError) Error() string { return "big error" }
