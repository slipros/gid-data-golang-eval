// Eval GID-211: the domain layer is not touched — in model an enum lives in model (GID-132).
package model

// --- Not-applicable class: a string enum in /domain/model — normal, no diagnostic ---

type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
)
