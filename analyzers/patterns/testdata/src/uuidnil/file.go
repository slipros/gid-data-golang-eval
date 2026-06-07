package uuidnil

import "github.com/gofrs/uuid"

// Позитив: сравнение UUID с uuid.UUID{} запрещено (== и !=).
func bad(id uuid.UUID) (bool, bool) {
	eq := id == uuid.UUID{} // want `GID-002: do not compare a UUID with uuid\.UUID\{\}\. Fix: replace "id == uuid\.UUID\{\}" with "id\.IsNil\(\)"\.`
	ne := id != uuid.UUID{} // want `GID-002: .* Fix: replace "id != uuid\.UUID\{\}" with "!id\.IsNil\(\)"\.`
	return eq, ne
}

// Негатив: канонический IsNil().
func good(id uuid.UUID) bool {
	return id.IsNil()
}

// Неприменимость: сравнение не-UUID типов.
func boundary(a, b int) bool {
	return a == b
}
