// Eval GID-233: model-layer enums (named string types with typed consts).
package enum

// Status is an enum per GID-123: a named string type with typed consts.
type Status string

const (
	StatusActive  Status = "active"
	StatusBlocked Status = "blocked"
)

// Kind is another enum declared in the SAME package as Status.
type Kind string

const KindPrimary Kind = "primary"

// Label is a named string type WITHOUT consts — not an enum (non-applicability).
type Label string

// KindFromStatus: a same-package enum→enum cast is allowed (boundary case) —
// inside the enum's own package family it is not a layer boundary.
func KindFromStatus(s Status) Kind {
	return Kind(s)
}
