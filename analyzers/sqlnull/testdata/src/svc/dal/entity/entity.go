// Eval GID-122 (sql.Null* instead of pointers in entity).
package entity

import (
	"database/sql"
	"time"
)

type MyType struct{ V int }

// --- Positive cases ---

type Snapshot struct {
	CompletedAt *time.Time // want `GID-122: a nullable entity field must use sql\.NullTime, not a pointer\. Fix: replace the pointer with it`
	Description *string    // want `GID-122: a nullable entity field must use sql\.NullString, not a pointer\. Fix: replace the pointer with it`
	FileCount   *int32     // want `GID-122: a nullable entity field must use sql\.NullInt32, not a pointer\. Fix: replace the pointer with it`
	Size        *int64     // want `GID-122: a nullable entity field must use sql\.NullInt64, not a pointer\. Fix: replace the pointer with it`

	// Boundary case: a non-standard type — the generic sql.Null[T].
	Custom *MyType // want `GID-122: a nullable entity field must use sql\.Null\[T\], not a pointer\. Fix: replace the pointer with it`
}

// --- Negative cases ---

type Job struct {
	CompletedAt sql.NullTime      `db:"completed_at"`
	Description sql.NullString    `db:"description"`
	FileCount   sql.NullInt32     `db:"file_count"`
	Custom      sql.Null[MyType]  `db:"custom_field"`
	CreatedAt   time.Time         `db:"created_at"` // not null — a regular type
}
