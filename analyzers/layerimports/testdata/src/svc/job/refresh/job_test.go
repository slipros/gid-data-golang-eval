// Non-applicability: a _test.go file is not judged. A test lives in the same
// package (GID-250), and a double of the repository has to speak that
// repository's types — the import mirrors what the package's own interfaces
// already declare, and the production file that declares them is judged as usual.
package refresh

import (
	"testing"

	"svc/dal/repository"
)

type fakeRepo struct{ snapshot *repository.Snapshot }

func TestJob(t *testing.T) {
	if (&fakeRepo{}).snapshot != nil {
		t.Fatal("unexpected snapshot")
	}
}
