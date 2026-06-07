package uuidversion

import "github.com/gofrs/uuid"

// Позитив: версии V1/V3/V4/V5/V6 запрещены.
func bad() {
	_, _ = uuid.NewV1()            // want `GID-003: UUIDs must be generated uniformly\. Fix: use uuid\.Must\(uuid\.NewV7\(\)\) instead of uuid\.NewV1\(\)\.`
	_ = uuid.NewV3(uuid.UUID{}, "") // want `GID-003: .* instead of uuid\.NewV3\(\)\.`
	_, _ = uuid.NewV4()            // want `GID-003: .* instead of uuid\.NewV4\(\)\.`
	_ = uuid.NewV5(uuid.UUID{}, "") // want `GID-003: .* instead of uuid\.NewV5\(\)\.`
	_, _ = uuid.NewV6()            // want `GID-003: .* instead of uuid\.NewV6\(\)\.`
}

// Негатив: V7 — канонический генератор.
func good() uuid.UUID {
	return uuid.Must(uuid.NewV7())
}
