package uuidnil

import "github.com/gofrs/uuid"

// Positive: comparing a UUID with uuid.UUID{} is forbidden (== and !=).
func bad(id uuid.UUID) (bool, bool) {
	eq := id == uuid.UUID{} // want `GID-002: do not compare a UUID with uuid\.UUID\{\}\. Fix: replace "id == uuid\.UUID\{\}" with "id\.IsNil\(\)"\.`
	ne := id != uuid.UUID{} // want `GID-002: .* Fix: replace "id != uuid\.UUID\{\}" with "!id\.IsNil\(\)"\.`
	return eq, ne
}

// Negative: the canonical IsNil().
func good(id uuid.UUID) bool {
	return id.IsNil()
}

// Not applicable: comparing non-UUID types.
func boundary(a, b int) bool {
	return a == b
}
