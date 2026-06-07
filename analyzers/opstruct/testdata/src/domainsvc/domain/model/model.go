// Eval GID-210: model-Create-структуры не содержат ID/CreatedAt/UpdatedAt.
package model

import "time"

// --- Позитивный класс: нарушения ---

// model-Create с генерируемыми полями — флагается каждое.
type CreateJob struct {
	Title     string
	ID        int       // want `GID-210: операционная структура "CreateJob" не должна содержать поле "ID" .* — убери его из Create`
	CreatedAt time.Time // want `GID-210: операционная структура "CreateJob" не должна содержать поле "CreatedAt" .* — убери его из Create`
	UpdatedAt time.Time // want `GID-210: операционная структура "CreateJob" не должна содержать поле "UpdatedAt" .* — убери его из Create`
}

// Несколько имён в одном поле — проверяется каждое.
type CreateStageInput struct {
	ID, UpdatedAt int // want `GID-210: операционная структура "CreateStageInput" не должна содержать поле "ID"` `GID-210: операционная структура "CreateStageInput" не должна содержать поле "UpdatedAt"`
}

// --- Негативный класс: чистый код проходит ---

// Чистая Create-структура — диагностики нет.
type CreateUser struct {
	Name  string
	Email string
}

// Обычная не-операционная структура (^Create[A-Z] не матчится) — ID/CreatedAt легитимны.
type Snapshot struct {
	ID        int
	CreatedAt time.Time
}

// --- Граничный класс ---

// CreatedBy не путается с CreatedAt — поле разрешено.
type CreateOrder struct {
	CreatedBy string
}

// CreatedSnapshot не матчится под ^Create[A-Z] (после Create идёт строчная d).
type CreatedSnapshot struct {
	ID        int
	CreatedAt time.Time
}

// Update-структуры правилом не задеваются.
type UpdateJob struct {
	ID        int
	UpdatedAt time.Time
}

// Голое имя Create без следующей заглавной — не операционная Create-структура.
type Create struct {
	ID int
}
