package uuidnilv5

import "github.com/gofrs/uuid/v5"

func nilSentinel(id uuid.UUID) bool {
	return id != uuid.Nil // want `GID-002: .*!id\.IsZero\(\)`
}

func zeroLiteral(id uuid.UUID) bool {
	return id == uuid.UUID{} // want `GID-002: .*id\.IsZero\(\)`
}

func correct(id uuid.UUID) bool { return id.IsZero() }

func differentUUID(a, b uuid.UUID) bool { return a == b }
