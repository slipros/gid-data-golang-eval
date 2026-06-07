// Eval GID-210: entity-Create содержит только поля INSERT (без UpdatedAt).
package entity

import "time"

// --- Позитивный класс: нарушение ---

// entity-Create с UpdatedAt — UpdatedAt флагается, ID и CreatedAt легитимны.
type CreateJob struct {
	ID        int
	CreatedAt time.Time
	Title     string
	UpdatedAt time.Time // want `GID-210: operational struct "CreateJob" must not contain field "UpdatedAt" .* Fix: remove it from Create`
}

// --- Негативный класс: чистый код проходит ---

// entity-Create с ID и CreatedAt, но без UpdatedAt — ок (это INSERT-поля).
type CreateUser struct {
	ID        int
	CreatedAt time.Time
	Name      string
}

// --- Граничный класс ---

// Update-структуры правилом не задеваются.
type UpdateJob struct {
	UpdatedAt time.Time
}

// CreatedSnapshot не матчится под ^Create[A-Z].
type CreatedSnapshot struct {
	UpdatedAt time.Time
}
