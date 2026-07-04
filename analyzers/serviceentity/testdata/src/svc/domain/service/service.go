// Eval for GID-236 (service-one-entity).
package service

import (
	"context"

	"svc/dal/repository"
)

// JobRepository — a dependency interface of another entity, used to test
// cross-entity injection.
type JobRepository interface {
	Job(ctx context.Context, id string) (string, error)
}

// SnapshotRepository — the service's own entity repository.
type SnapshotRepository interface {
	Snapshot(ctx context.Context, id string) (string, error)
}

// SnapshotValidator — an interface without the Repository suffix: not a
// repository dependency at all.
type SnapshotValidator interface {
	Validate(ctx context.Context, id string) error
}

// SnapshotFileRepository — a repository interface whose entity
// ("SnapshotFile") merely shares a prefix with "Snapshot" but differs from it.
type SnapshotFileRepository interface {
	File(ctx context.Context, id string) (string, error)
}

// Options — service settings, not an entity struct.
type Options struct {
	MaxSize int
}

// SnapshotOptions — an *Options type is skipped entirely, even when it
// holds a foreign repository.
type SnapshotOptions struct {
	Jobs JobRepository
}

// --- Positive: a named field injecting a foreign entity's repository ---

type Upload struct {
	jobs JobRepository // want `GID-236: service "Upload" uses repository "JobRepository" of another entity\. Fix: a service works with exactly one entity — orchestrate several entities in usecase \(or //nolint:gidserviceentity when explicitly intended\)`
}

// --- Positive: an embedded foreign-entity repository ---

type Delivery struct {
	JobRepository // want `GID-236: service "Delivery" uses repository "JobRepository" of another entity`
}

// --- Negative: the service's own repository, a non-repository interface and
// Options are the norm. Boundary: a repository whose entity only shares a
// prefix with the owner ("SnapshotFile" vs "Snapshot") is still foreign ---

type Snapshot struct {
	repository SnapshotRepository
	validator  SnapshotValidator
	opts       Options
	files      SnapshotFileRepository // want `GID-236: service "Snapshot" uses repository "SnapshotFileRepository" of another entity`
}

// --- Negative: an interface declared in another package is out of the
// GID-134/GID-236 same-package scope ---

type Report struct {
	files repository.FileRepository
}
