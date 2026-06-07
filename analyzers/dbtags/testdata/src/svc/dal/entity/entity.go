// Eval GID-125 (db-теги в entity).
package entity

import "time"

// --- Позитивные кейсы ---

type Snapshot struct {
	ID        string    `db:"id"`
	Name      string    // want `GID-125: field Snapshot\.Name has no mapping tag \(db\)\. Fix: add a tag so entity-to-column mapping is explicit`
	CreatedAt time.Time `json:"created_at"` // want `GID-125: field Snapshot\.CreatedAt has no mapping tag \(db\)`
}

// --- Негативные кейсы ---

type Job struct {
	ID        string    `db:"id"`
	Status    string    `db:"status"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// Неприменимость: приватные поля не маппятся напрямую.
type cursor struct {
	offset int
}
