// Eval for GID-194: the model layer — package-level constants are legal.
package model

const (
	StatusActive  = "active"
	StatusDeleted = "deleted"
)

const maxNameLen = 100

var _ = maxNameLen
