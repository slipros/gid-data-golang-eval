// Eval GID-168 (a ban on db tags in /domain/**).
package model

import "time"

// --- Positive cases: a db tag in domain is a violation ---

type Snapshot struct {
	ID        string    `db:"id"`                            // want `GID-168: field Snapshot\.ID has a "db" tag in the domain layer\. Fix: keep db mapping in /dal/entity`
	Name      string    `db:"name" json:"name"`              // want `GID-168: field Snapshot\.Name has a "db" tag in the domain layer`
	CreatedAt time.Time `json:"created_at" db:"created_at"`  // want `GID-168: field Snapshot\.CreatedAt has a "db" tag in the domain layer`
}

// Positive: a private field with a db tag is flagged too.
type cursor struct {
	offset int `db:"offset"` // want `GID-168: field cursor\.offset has a "db" tag in the domain layer`
}

// --- Edge cases ---

// Edge: an embedded field with a db tag — flagged (the name = the type name).
type WithEmbedded struct {
	Snapshot `db:"snapshot"` // want `GID-168: field WithEmbedded\.Snapshot has a "db" tag in the domain layer`
	Extra    string
}

// Edge: a ch tag with the default settings (["db"]) — NOT flagged.
type Metric struct {
	ID    string `ch:"id"`
	Value int64  `ch:"value"`
}

// --- Negative cases: no mapping tags — clean ---

type Job struct {
	ID     string
	Status string `json:"status"`
	Title  string `json:"title" validate:"required"`
}
