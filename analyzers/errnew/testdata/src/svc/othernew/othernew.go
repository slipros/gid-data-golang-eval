// Package othernew contains a local New function unrelated to pkg/errors.
package othernew

// New — a same-named function from another package; rule GID-136 does not touch it.
func New(message string) error { return nil }
