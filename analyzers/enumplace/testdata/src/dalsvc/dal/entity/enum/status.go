// Eval GID-211: the canonical place for an enum is /dal/entity/enum, no diagnostic.
package enum

// --- Negative class: a string enum in its proper place — ok ---

type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
)
