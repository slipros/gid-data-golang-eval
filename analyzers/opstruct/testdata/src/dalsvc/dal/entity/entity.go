// Eval GID-210: an entity Create contains only INSERT fields (no UpdatedAt).
package entity

import "time"

// --- Positive class: violation ---

// An entity Create with UpdatedAt — UpdatedAt is flagged, ID and CreatedAt are legitimate.
type CreateJob struct {
	ID        int
	CreatedAt time.Time
	Title     string
	UpdatedAt time.Time // want `GID-210: operational struct "CreateJob" must not contain field "UpdatedAt" .* Fix: remove it from Create`
}

// --- Negative class: clean code passes ---

// An entity Create with ID and CreatedAt but without UpdatedAt — ok (those are INSERT fields).
type CreateUser struct {
	ID        int
	CreatedAt time.Time
	Name      string
}

// --- Boundary class ---

// Update structs are not affected by the rule.
type UpdateJob struct {
	UpdatedAt time.Time
}

// CreatedSnapshot does not match ^Create[A-Z].
type CreatedSnapshot struct {
	UpdatedAt time.Time
}
