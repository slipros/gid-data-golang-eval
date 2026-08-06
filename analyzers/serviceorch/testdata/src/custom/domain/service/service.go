// Eval of the GID-260 settings: settings.suffixes adds "Store" on top of the
// default "Repository"; settings.exclude clears a whole struct ("LegacyWriter")
// and a single field ("Mixed.tx").
package service

import (
	"context"

	"custom/domain/model"
)

type SnapshotRepository interface {
	Snapshot(ctx context.Context, id string) (string, error)
}

type CoreSnapshotStore interface {
	CoreSnapshot(ctx context.Context, id string) (string, error)
}

// --- Positive: the custom "Store" suffix counts as a second source ---

type Integration struct { // want `GID-260: service "Integration" depends on 2 repositories \(SnapshotRepository, CoreSnapshotStore\)`
	repo  SnapshotRepository
	store CoreSnapshotStore
}

// --- Non-applicability: the whole struct is on the exclusion list ---

type LegacyWriter struct {
	tx    model.InTransactionFunc
	repo  SnapshotRepository
	store CoreSnapshotStore
}

// --- Non-applicability: only the tx field is excluded, the pair of sources is
// still reported ---

type Mixed struct { // want `GID-260: service "Mixed" depends on 2 repositories`
	tx    model.InTransactionFunc
	repo  SnapshotRepository
	store CoreSnapshotStore
}
