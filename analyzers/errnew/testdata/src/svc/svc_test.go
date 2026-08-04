// Non-applicability: a _test.go file is not judged. A test lives in the same
// package (GID-250), so hoisting the errors.New of a table case to a
// package-level var would declare a fixture in the production namespace.
package svc

import (
	"testing"

	"github.com/pkg/errors"
)

func TestCases(t *testing.T) {
	cases := []struct {
		name string
		err  error
	}{
		{name: "down", err: errors.New("repo down")},
		{name: "timeout", err: errors.New("timeout")},
	}
	for _, c := range cases {
		if c.err == nil {
			t.Fatal(c.name)
		}
	}
}
