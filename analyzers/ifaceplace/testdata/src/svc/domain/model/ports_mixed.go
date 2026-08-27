// Eval for GID-271, non-applicability class: a file with a struct next to the
// interface is not a port file (top-level declarations are not only
// interfaces) — the file is the consumer's file already, nothing to move.
package model

// Ref is declared next to its only consumer — not a port file.
type Ref interface {
	ID() string
}

// RefHolder keeps the interface at its side in the same file.
type RefHolder struct {
	ref Ref
}
