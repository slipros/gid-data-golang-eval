// Eval GID-125 (db tags in entity).
package entity

import "time"

// --- Positive cases ---

type Snapshot struct {
	ID        string    `db:"id"`
	Name      string    // want `GID-125: field Snapshot\.Name has no mapping tag \(db\)\. Fix: add a tag so entity-to-column mapping is explicit`
	CreatedAt time.Time `json:"created_at"` // want `GID-125: field Snapshot\.CreatedAt has no mapping tag \(db\)`
}

// --- Negative cases ---

type Job struct {
	ID        string    `db:"id"`
	Status    string    `db:"status"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// Not applicable: private fields are not mapped directly.
type cursor struct {
	offset int
}
