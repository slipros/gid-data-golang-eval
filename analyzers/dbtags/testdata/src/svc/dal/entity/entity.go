// Eval GID-125 (db-теги в entity).
package entity

import "time"

// --- Позитивные кейсы ---

type Snapshot struct {
	ID        string    `db:"id"`
	Name      string    // want `GID-125: поле Snapshot\.Name без тега маппинга \(db\) — соответствие entity колонкам БД явное`
	CreatedAt time.Time `json:"created_at"` // want `GID-125: поле Snapshot\.CreatedAt без тега маппинга \(db\)`
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
