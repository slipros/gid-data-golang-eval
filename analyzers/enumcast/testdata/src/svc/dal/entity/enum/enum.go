// Eval GID-233: entity-layer enums (named string types with typed consts).
package enum

// Status is an enum per GID-123: a named string type with typed consts.
type Status string

const (
	StatusActive  Status = "active"
	StatusBlocked Status = "blocked"
)

// Raw is a named string type WITHOUT consts — not an enum (non-applicability).
type Raw string
