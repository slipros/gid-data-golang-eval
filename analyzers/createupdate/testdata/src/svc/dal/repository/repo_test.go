// Non-applicability: a _test.go file is not judged. A test lives in the same
// package (GID-250), and a double implementing the interface under test copies
// its signatures — the shape of the method is the interface's, not the double's.
package repository

import (
	"context"
	"testing"
)

type fakeJobRepo struct{ snapshot Snapshot }

// CreateJob returns data because the interface it doubles says so.
func (f *fakeJobRepo) CreateJob(ctx context.Context, name string) (Snapshot, error) {
	return f.snapshot, nil
}

func TestCreateJob(t *testing.T) {
	if _, err := (&fakeJobRepo{}).CreateJob(context.Background(), "n"); err != nil {
		t.Fatal(err)
	}
}
