// Non-applicability: a _test.go file is not judged. A test lives in the same
// package (GID-250), and its sentinel belongs to the test — moving it to
// /domain/model would put a fixture into production code.
package service

import (
	"testing"

	"github.com/pkg/errors"
)

// errRepoDown is handed to a fake repository to check the error is propagated.
var errRepoDown = errors.New("repo down")

func TestSnapshot(t *testing.T) {
	if err := errors.New("ad-hoc"); err == nil {
		t.Fatal(errRepoDown)
	}
}
