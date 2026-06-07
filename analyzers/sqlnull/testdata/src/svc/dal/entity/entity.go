// Eval GID-122 (sql.Null* вместо указателей в entity).
package entity

import (
	"database/sql"
	"time"
)

type MyType struct{ V int }

// --- Позитивные кейсы ---

type Snapshot struct {
	CompletedAt *time.Time // want `GID-122: nullable-поле entity описывается типом sql\.NullTime, не указателем`
	Description *string    // want `GID-122: nullable-поле entity описывается типом sql\.NullString, не указателем`
	FileCount   *int32     // want `GID-122: nullable-поле entity описывается типом sql\.NullInt32, не указателем`
	Size        *int64     // want `GID-122: nullable-поле entity описывается типом sql\.NullInt64, не указателем`

	// Граничный кейс: нестандартный тип — обобщённый sql.Null[T].
	Custom *MyType // want `GID-122: nullable-поле entity описывается типом sql\.Null\[T\], не указателем`
}

// --- Негативные кейсы ---

type Job struct {
	CompletedAt sql.NullTime      `db:"completed_at"`
	Description sql.NullString    `db:"description"`
	FileCount   sql.NullInt32     `db:"file_count"`
	Custom      sql.Null[MyType]  `db:"custom_field"`
	CreatedAt   time.Time         `db:"created_at"` // not null — обычный тип
}
