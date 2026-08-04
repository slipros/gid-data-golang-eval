// Non-applicability: a _test.go file is not judged. Tests live in the same
// package (GID-250), and a test double implementing the interface under test
// copies the method names from that interface — renaming them is impossible.
package repository

import (
	"context"
	"testing"
)

// fakeJobRepo implements the same interface as Job: the method names come from
// the interface, not from the double's own type name.
type fakeJobRepo struct{ snapshot Snapshot }

func (f *fakeJobRepo) Job(ctx context.Context, id string) (Snapshot, error) {
	return f.snapshot, nil
}

// A verb method, a List prefix and a ByID suffix — all silent in a test file.
func (f *fakeJobRepo) Fetch(ctx context.Context) (Snapshot, error) {
	return f.snapshot, nil
}

func (f *fakeJobRepo) ListJobs(ctx context.Context) ([]Snapshot, error) {
	return nil, nil
}

func (f *fakeJobRepo) JobByID(ctx context.Context, id string) (Snapshot, error) {
	return f.snapshot, nil
}

func TestJob(t *testing.T) {
	if _, err := (&fakeJobRepo{}).Job(context.Background(), "id"); err != nil {
		t.Fatal(err)
	}
}
