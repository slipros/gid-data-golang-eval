// Eval для GID-002 (no-uuid-empty-compare).
package nouuidcompare

import "github.com/gofrs/uuid"

type other [16]byte

// --- Позитивные кейсы: нарушение ловится ---

func badEq(id uuid.UUID) bool {
	return id == uuid.UUID{} // want `GID-002: используйте id\.IsNil\(\) вместо сравнения с uuid\.UUID\{\}`
}

func badNeq(id uuid.UUID) bool {
	return id != uuid.UUID{} // want `GID-002: используйте !id\.IsNil\(\) вместо сравнения с uuid\.UUID\{\}`
}

// Граничный кейс: сравнение поля структуры.
type job struct {
	parentID uuid.UUID
}

func badField(j *job) bool {
	return j.parentID == uuid.UUID{} // want `GID-002: используйте j\.parentID\.IsNil\(\) вместо сравнения с uuid\.UUID\{\}`
}

// --- Негативные кейсы: чистый код проходит ---

func good(id uuid.UUID) bool {
	return id.IsNil()
}

func goodNot(id uuid.UUID) bool {
	return !id.IsNil()
}

// --- Неприменимость: сравнение других типов с composite literal ---

func notApplicable(o other) bool {
	return o == other{}
}
