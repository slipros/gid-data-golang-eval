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

// Boundary: inside an if condition the literal must be parenthesised to parse —
// the parentheses do not make the comparison legal.
type request struct{ DatasetID uuid.UUID }

func badInCondition(in request) bool {
	if in.DatasetID != (uuid.UUID{}) { // want `GID-002: .* Fix: replace "in\.DatasetID != uuid\.UUID\{\}" with "!in\.DatasetID\.IsNil\(\)"\.`
		return true
	}
	if (uuid.UUID{}) == in.DatasetID { // want `GID-002: .* Fix: replace "in\.DatasetID == uuid\.UUID\{\}" with "in\.DatasetID\.IsNil\(\)"\.`
		return false
	}
	return false
}
