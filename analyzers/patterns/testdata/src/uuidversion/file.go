package uuidversion

import "github.com/gofrs/uuid"

// Positive: versions V1/V3/V4/V5/V6 are forbidden.
func bad() {
	_, _ = uuid.NewV1()            // want `GID-003: UUIDs must be generated uniformly\. Fix: use uuid\.Must\(uuid\.NewV7\(\)\) instead of uuid\.NewV1\(\)\.`
	_ = uuid.NewV3(uuid.UUID{}, "") // want `GID-003: .* instead of uuid\.NewV3\(\)\.`
	_, _ = uuid.NewV4()            // want `GID-003: .* instead of uuid\.NewV4\(\)\.`
	_ = uuid.NewV5(uuid.UUID{}, "") // want `GID-003: .* instead of uuid\.NewV5\(\)\.`
	_, _ = uuid.NewV6()            // want `GID-003: .* instead of uuid\.NewV6\(\)\.`
}

// Positive: NewV7 must be wrapped in uuid.Must, not error-handled.
func badMust() (uuid.UUID, error) {
	return uuid.NewV7() // want `GID-003: UUIDs must be generated via uuid\.Must\. Fix: use uuid\.Must\(uuid\.NewV7\(\)\) instead of handling the error\.`
}

func badMustVar() uuid.UUID {
	id, _ := uuid.NewV7() // want `GID-003: UUIDs must be generated via uuid\.Must\. Fix: use uuid\.Must\(uuid\.NewV7\(\)\) instead of handling the error\.`
	return id
}

// Positive: wrapping a banned version in uuid.Must does not launder it —
// only V7 is the canonical generator.
func badWrapped() uuid.UUID {
	return uuid.Must(uuid.NewV4()) // want `GID-003: .* instead of uuid\.NewV4\(\)\.`
}

// Negative: V7 is the canonical generator.
func good() uuid.UUID {
	return uuid.Must(uuid.NewV7())
}
