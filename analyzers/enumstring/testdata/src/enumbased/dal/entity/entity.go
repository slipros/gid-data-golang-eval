// Eval GID-123 в /dal/entity.
package entity

// --- Позитив: int-enum с ≥2 const-значений ---

type Status int // want `GID-123: enum Status must be based on string, not int`

const (
	StatusActive   Status = 1
	StatusInactive Status = 2
)

// --- Позитив: alias на string ---

type Code = string // want `GID-123: enum Code must be a named type, not an alias`

// --- Негатив: правильный string-enum ---

type Kind string

const (
	KindA Kind = "a"
	KindB Kind = "b"
)

// --- Граничный: одиночная const именованного int-типа — не enum ---

type Priority int

const DefaultPriority Priority = 5
