// Eval GID-123 in /dal/entity.
package entity

// --- Positive: int-enum with ≥2 const values ---

type Status int // want `GID-123: enum Status must be based on string, not int`

const (
	StatusActive   Status = 1
	StatusInactive Status = 2
)

// --- Positive: alias to string ---

type Code = string // want `GID-123: enum Code must be a named type, not an alias`

// --- Negative: a correct string-enum ---

type Kind string

const (
	KindA Kind = "a"
	KindB Kind = "b"
)

// --- Edge case: a single const of a named int type — not an enum ---

type Priority int

const DefaultPriority Priority = 5
