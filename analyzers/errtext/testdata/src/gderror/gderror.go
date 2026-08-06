// Package gderror stands in for a foreign error constructor — a package the
// rule does not know. GID-146 keeps production code on pkg/errors anyway.
package gderror

import stderrors "errors"

// New builds an error from a message.
func New(message string) error { return stderrors.New(message) }
