// Eval GID-120/121 (pointers in model).
package model

import (
	"time"

	"github.com/gofrs/uuid"
)

type SnapshotStatus string

// Weight is a named type based on int — a boundary case: the ban on
// pointers to simple types applies by the *underlying* basic kind, not just
// to the bare int/string types.
type Weight int

// --- Positive cases ---

type Snapshot struct {
	ParentID    *uuid.UUID      // want `GID-120: \*uuid\.UUID is forbidden\. Fix: use uuid\.UUID and check emptiness with IsNil\(\)`
	CompletedAt *time.Time      // want `GID-121: \*time\.Time is unnecessary here\. Fix: use time\.Time and check absence with t\.IsZero\(\)`
	Description *string         // want `GID-121: a pointer to a simple type is unnecessary here\. Fix: use the value and check the zero value`
	Status      *SnapshotStatus // want `GID-121: a pointer to a simple type is unnecessary here`
	Count       *int            // want `GID-121: a pointer to a simple type is unnecessary here`
	Ratio       *float64        // want `GID-121: a pointer to a simple type is unnecessary here`
	Priority    *Weight         // want `GID-121: a pointer to a simple type is unnecessary here`
}

// Boundary case: *uuid.UUID in a signature is also a GID-120 violation.
func Lookup(id *uuid.UUID) bool { // want `GID-120: \*uuid\.UUID is forbidden\. Fix: use uuid\.UUID and check emptiness with IsNil\(\)`
	return id != nil
}

// --- Negative cases ---

type Job struct {
	ID          uuid.UUID
	CompletedAt time.Time
	Description string
	Status      SnapshotStatus
	Count       int
	Enabled     *bool // the pointer is justified: false is a valid value
	Parent      *Job  // a nested struct — a pointer is allowed
}

// Not applicable: a dereference is not a type.
func deref(id *bool) bool { return *id }
