// Eval GID-210: model Create structs do not contain ID/CreatedAt/UpdatedAt.
package model

import "time"

// --- Positive class: violations ---

// A model Create with generated fields — each one is flagged.
type CreateJob struct {
	Title     string
	ID        int       // want `GID-210: operational struct "CreateJob" must not contain field "ID" .* Fix: remove it from Create`
	CreatedAt time.Time // want `GID-210: operational struct "CreateJob" must not contain field "CreatedAt" .* Fix: remove it from Create`
	UpdatedAt time.Time // want `GID-210: operational struct "CreateJob" must not contain field "UpdatedAt" .* Fix: remove it from Create`
}

// Several names in one field — each one is checked.
type CreateStageInput struct {
	ID, UpdatedAt int // want `GID-210: operational struct "CreateStageInput" must not contain field "ID"` `GID-210: operational struct "CreateStageInput" must not contain field "UpdatedAt"`
}

// --- Negative class: clean code passes ---

// A clean Create struct — no diagnostic.
type CreateUser struct {
	Name  string
	Email string
}

// An ordinary non-operational struct (^Create[A-Z] does not match) — ID/CreatedAt are legitimate.
type Snapshot struct {
	ID        int
	CreatedAt time.Time
}

// --- Boundary class ---

// CreatedBy is not confused with CreatedAt — the field is allowed.
type CreateOrder struct {
	CreatedBy string
}

// CreatedSnapshot does not match ^Create[A-Z] (a lowercase d follows Create).
type CreatedSnapshot struct {
	ID        int
	CreatedAt time.Time
}

// Update structs are not affected by the rule.
type UpdateJob struct {
	ID        int
	UpdatedAt time.Time
}

// The bare name Create without a following capital is not an operational Create struct.
type Create struct {
	ID int
}
