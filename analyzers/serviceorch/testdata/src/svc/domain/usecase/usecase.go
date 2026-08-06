// Non-applicability: a usecase is where orchestration belongs — several
// repositories and the transaction are the norm here, and the rule only judges
// /domain/service.
package usecase

import (
	"context"

	"svc/domain/model"
)

type SnapshotRepository interface {
	Snapshot(ctx context.Context, id string) (string, error)
}

type CoreSnapshotRepository interface {
	CoreSnapshot(ctx context.Context, id string) (string, error)
}

type Integration struct {
	tx       model.InTransactionFunc
	coreRepo CoreSnapshotRepository
	repo     SnapshotRepository
}
