// Non-applicability: a _test.go file is not judged. Tests live in the same
// package (GID-250), and a harness holding the service under test next to its
// fixtures is composition of the test, not a service-on-service dependency.
package service

import "testing"

// snapshotHarness bundles the service under test with a second service it is
// compared against — legal only because this is a test file.
type snapshotHarness struct {
	svc   *Snapshot
	other *Job
}

func TestSnapshotHarness(t *testing.T) {
	h := snapshotHarness{svc: &Snapshot{}, other: &Job{}}
	if h.svc == nil || h.other == nil {
		t.Fatal("nil service")
	}
}
