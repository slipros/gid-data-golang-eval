// Eval GID-123 scope extended to /event/dto: DTO types follow the same
// enum rules as /domain/model and /dal/entity.
package dto

// --- Positive: alias to a basic string ---

type StatusDTO = string // want `GID-123: enum StatusDTO must be a named type, not an alias`

// --- Positive: int-enum with ≥2 const values ---

type Kind int // want `GID-123: enum Kind must be based on string, not int`

const (
	KindA Kind = 1
	KindB Kind = 2
)

// --- Positive: a group of untyped string constants ---

const (
	RoleAdmin = "admin" // want `GID-123: a group of string constants\. Fix: declare a named string type`
	RoleUser  = "user"
)

// --- Negative: a correct string-based enum ---

type EventType string

const (
	EventTypeCreated EventType = "created"
	EventTypeDeleted EventType = "deleted"
)
