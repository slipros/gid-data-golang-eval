// Eval для GID-194: model-слой — package-level константы легальны.
package model

const (
	StatusActive  = "active"
	StatusDeleted = "deleted"
)

const maxNameLen = 100

var _ = maxNameLen
