// Eval GID-211: an enum in the root of /dal/entity instead of /dal/entity/enum.
package entity

// --- Positive class: a string enum with const in /dal/entity — a violation ---

type Status string // want `GID-211: enum Status must live in /dal/entity/enum \(one file named after the entity\)\. Fix: move it there`

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
)

// --- Negative class: a string type without const — not an enum, not flagged ---

type RawJSON string

// --- Boundary class: an alias to string with const — GID-123 zone, not GID-211 ---

type Code = string

const (
	CodeA Code = "a"
	CodeB Code = "b"
)

// --- Boundary class: a named int type with const — not a string enum ---

type Priority int

const (
	PriorityLow  Priority = 1
	PriorityHigh Priority = 2
)

// --- An ordinary entity without an enum — left alone ---

type Job struct {
	ID string
}
