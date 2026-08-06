// Eval for GID-260 (service-no-orchestration).
package service

import (
	"context"

	"svc/domain/model"
)

// SnapshotRepository — the service's own entity repository.
type SnapshotRepository interface {
	Snapshot(ctx context.Context, id string) (string, error)
}

// CoreSnapshotRepository — a second source: the core record of the same
// business object, living in another layer's data store.
type CoreSnapshotRepository interface {
	CoreSnapshot(ctx context.Context, id string) (string, error)
}

// SnapshotValidator — an interface without the repository suffix: not a source.
type SnapshotValidator interface {
	Validate(ctx context.Context, id string) error
}

// Options — service settings, not a service struct.
type Options struct {
	MaxSize int
}

// SnapshotOptions — an *Options type is skipped entirely, even holding two
// repositories.
type SnapshotOptions struct {
	repo     SnapshotRepository
	coreRepo CoreSnapshotRepository
}

// --- Positive: two repositories in one service ---

type Integration struct { // want `GID-260: service "Integration" depends on 2 repositories \(CoreSnapshotRepository, SnapshotRepository\) — a service is one entity and its repository\. Fix: split it into a service per entity and compose them in a usecase \(or //nolint:gidserviceorch when explicitly intended\)`
	coreRepo CoreSnapshotRepository
	repo     SnapshotRepository
}

// --- Positive: a transaction held by a service ---

type Writer struct {
	tx   model.InTransactionFunc // want `GID-260: service "Writer" holds a transaction — a service does not coordinate several writes\. Fix: keep the transaction in a usecase, which calls the services it composes \(or //nolint:gidserviceorch when explicitly intended\)`
	repo SnapshotRepository
}

// --- Positive: the incident shape — both markers at once ---

type CoreAndTypeWriter struct { // want `GID-260: service "CoreAndTypeWriter" depends on 2 repositories`
	tx       model.InTransactionFunc // want `GID-260: service "CoreAndTypeWriter" holds a transaction`
	coreRepo CoreSnapshotRepository
	repo     SnapshotRepository
}

// --- Negative: one repository, a validator and options are the norm ---

type Snapshot struct {
	repo      SnapshotRepository
	validator SnapshotValidator
	opts      Options
}

// --- Boundary: a plain callback field is not a transaction runner — its
// parameter is not a func returning error ---

type Notifier struct {
	repo   SnapshotRepository
	notify func(ctx context.Context, id string) error
}

// --- Boundary: a func field returning something besides a lone error is not a
// transaction runner either ---

type Loader struct {
	repo SnapshotRepository
	load func(ctx context.Context, fn func(ctx context.Context) error) (string, error)
}
