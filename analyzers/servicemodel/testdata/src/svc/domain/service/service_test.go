// Non-applicability: a _test.go file is not judged. Tests live in the same
// package (GID-250), and a test double of the repository speaks the repository
// contract — entity in its signatures is correct, not a leak of the service API.
package service

import (
	"context"
	"testing"

	"svc/dal/entity"
)

// fakeSnapshotRepo doubles SnapshotRepository: it takes and returns entity
// because the repository contract says so.
type fakeSnapshotRepo struct{ snapshot entity.Snapshot }

func (f *fakeSnapshotRepo) Snapshot(ctx context.Context, id string) (entity.Snapshot, error) {
	return f.snapshot, nil
}

func (f *fakeSnapshotRepo) CreateSnapshot(ctx context.Context, in *entity.CreateSnapshot) error {
	return nil
}

func TestSnapshot(t *testing.T) {
	if _, err := (&fakeSnapshotRepo{}).Snapshot(context.Background(), "id"); err != nil {
		t.Fatal(err)
	}
}
