// Eval of GID-173 — /dal/repository: the bare role Connection + boundary cases.
package repository

import "context"

// --- Positive case: a bare role ---

type Connection interface { // want `GID-173: interface "Connection" must be named with an entity prefix\. Fix: e\.g\. HelloRepository`
	Ping(ctx context.Context) error
}

// --- Boundary case: a struct type with a role name — not an interface ---

type Repository struct {
	conn Connection
}

// --- Negative case: an interface with an entity prefix ---

type SnapshotConnection interface {
	Ping(ctx context.Context) error
}
