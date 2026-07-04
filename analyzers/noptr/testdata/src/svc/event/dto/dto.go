// Eval GID-121 scope extended to /event/dto: DTO types follow the same
// pointer-to-simple-type rules as /domain/model (event.md).
package dto

import "time"

// --- Positive cases ---

type SnapshotDTO struct {
	CompletedAt *time.Time // want `GID-121: \*time\.Time is unnecessary here\. Fix: use time\.Time and check absence with t\.IsZero\(\)`
	Count       *int       // want `GID-121: a pointer to a simple type is unnecessary here`
	Ratio       *float64   // want `GID-121: a pointer to a simple type is unnecessary here`
}

// --- Negative cases ---

type EventDTO struct {
	Enabled *bool        // the pointer is justified: false is a valid value
	Parent  *SnapshotDTO // a nested struct — a pointer is allowed
}
