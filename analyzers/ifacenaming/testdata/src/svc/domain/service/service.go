// Eval GID-173 (префикс сущности у интерфейсов зависимостей) — /domain/service.
package service

import "context"

// --- Позитивный кейс: голая роль ---

type Repository interface { // want `GID-173: интерфейс "Repository" именуется с префиксом сущности \(например, HelloRepository\)`
	Hello(ctx context.Context) error
}

// --- Негативные кейсы: имя с префиксом сущности ---

type HelloRepository interface {
	Hello(ctx context.Context) error
}

type SnapshotConnection interface {
	Ping(ctx context.Context) error
}

// --- Граничный кейс: имя содержит роль суффиксом, но не равно точно ---

type RepositoryFactory interface {
	New() HelloRepository
}
