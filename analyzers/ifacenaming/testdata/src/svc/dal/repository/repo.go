// Eval GID-173 — /dal/repository: голая роль Connection + граничные кейсы.
package repository

import "context"

// --- Позитивный кейс: голая роль ---

type Connection interface { // want `GID-173: interface "Connection" must be named with an entity prefix\. Fix: e\.g\. HelloRepository`
	Ping(ctx context.Context) error
}

// --- Граничный кейс: тип-структура с именем роли — не интерфейс ---

type Repository struct {
	conn Connection
}

// --- Негативный кейс: интерфейс с префиксом сущности ---

type SnapshotConnection interface {
	Ping(ctx context.Context) error
}
