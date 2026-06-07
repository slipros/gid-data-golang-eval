// Eval для GID-003 (uuid-only-v7).
package uuidonlyv7

import "github.com/gofrs/uuid"

// --- Позитивные кейсы: нарушение ловится ---

func badV4() uuid.UUID {
	return uuid.Must(uuid.NewV4()) // want `GID-003: UUID генерируются единообразно — uuid\.Must\(uuid\.NewV7\(\)\)`
}

func badV1() (uuid.UUID, error) {
	return uuid.NewV1() // want `GID-003: UUID генерируются единообразно — uuid\.Must\(uuid\.NewV7\(\)\)`
}

// Граничный кейс: генераторы с аргументами.
func badV5(ns uuid.UUID) uuid.UUID {
	return uuid.NewV5(ns, "name") // want `GID-003: UUID генерируются единообразно — uuid\.Must\(uuid\.NewV7\(\)\)`
}

// --- Негативные кейсы: чистый код проходит ---

func good() uuid.UUID {
	return uuid.Must(uuid.NewV7())
}

// --- Неприменимость: парсинг — не генерация ---

func notApplicable(s string) (uuid.UUID, error) {
	return uuid.FromString(s)
}
