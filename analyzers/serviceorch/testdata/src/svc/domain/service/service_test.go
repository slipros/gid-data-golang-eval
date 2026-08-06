// Non-applicability: a _test.go file is not judged — a harness bundling
// several doubles is composition of the test, not of a service (GID-250,
// same relaxation as GID-148/236).
package service

import (
	"context"

	"svc/domain/model"
)

type fixture struct {
	tx       model.InTransactionFunc
	coreRepo CoreSnapshotRepository
	repo     SnapshotRepository
}

func newFixture() *fixture {
	return &fixture{
		tx: func(ctx context.Context, fn func(ctx context.Context) error) error { return fn(ctx) },
	}
}
