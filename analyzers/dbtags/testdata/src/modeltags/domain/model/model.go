// Eval GID-168 (запрет db-тегов в /domain/**).
package model

import "time"

// --- Позитивные кейсы: db-тег в domain — нарушение ---

type Snapshot struct {
	ID        string    `db:"id"`                            // want `GID-168: field Snapshot\.ID has a "db" tag in the domain layer\. Fix: keep db mapping in /dal/entity`
	Name      string    `db:"name" json:"name"`              // want `GID-168: field Snapshot\.Name has a "db" tag in the domain layer`
	CreatedAt time.Time `json:"created_at" db:"created_at"`  // want `GID-168: field Snapshot\.CreatedAt has a "db" tag in the domain layer`
}

// Позитив: приватное поле с db-тегом тоже флагуется.
type cursor struct {
	offset int `db:"offset"` // want `GID-168: field cursor\.offset has a "db" tag in the domain layer`
}

// --- Граничные кейсы ---

// Граничный: embedded-поле с db-тегом — флагуем (имя = имя типа).
type WithEmbedded struct {
	Snapshot `db:"snapshot"` // want `GID-168: field WithEmbedded\.Snapshot has a "db" tag in the domain layer`
	Extra    string
}

// Граничный: ch-тег при настройках по умолчанию (["db"]) — НЕ флагуем.
type Metric struct {
	ID    string `ch:"id"`
	Value int64  `ch:"value"`
}

// --- Негативные кейсы: тегов маппинга нет — чисто ---

type Job struct {
	ID     string
	Status string `json:"status"`
	Title  string `json:"title" validate:"required"`
}
