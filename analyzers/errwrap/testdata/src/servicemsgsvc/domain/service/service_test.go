// Non-applicability: a _test.go file is not judged. A test lives in the same
// package (GID-250) and builds the very error shape the production code would
// produce — WithMessage in a fixture reproduces the input, it does not add a
// message on the way out of a service.
package service

import (
	"testing"

	"github.com/pkg/errors"
)

// fakeConverter returns the error a real converter would hand the service.
type fakeConverter struct{ err error }

func TestService(t *testing.T) {
	f := fakeConverter{err: errors.WithMessage(ErrConverted, "parse error")}
	if f.err == nil {
		t.Fatal("no error")
	}
}
