// Eval of GID-173 (an entity prefix on dependency interfaces) — /domain/service.
package service

import "context"

// --- Positive case: a bare role ---

type Repository interface { // want `GID-173: interface "Repository" must be named with an entity prefix\. Fix: e\.g\. HelloRepository`
	Hello(ctx context.Context) error
}

// --- Negative cases: a name with an entity prefix ---

type HelloRepository interface {
	Hello(ctx context.Context) error
}

type SnapshotConnection interface {
	Ping(ctx context.Context) error
}

// --- Boundary case: the name contains the role as a suffix but is not an exact match ---

type RepositoryFactory interface {
	New() HelloRepository
}
