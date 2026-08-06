// Non-applicability of the import ban: a _test.go file may hold the std
// package. A test of an error predicate takes its expected value from the std
// function on purpose — checking pkg/errors with pkg/errors would prove nothing.
//
// Creating errors through the std package is judged in a test all the same, so
// the constructor calls below are reported at their call site.
package onlypkgerrors

import (
	stderrors "errors"
	"testing"
)

var errTestSentinel = stderrors.New("test sentinel") // want `GID-146: errors\.New is forbidden\. Fix: use only github\.com/pkg/errors for errors`

func TestPredicate(t *testing.T) {
	want := stderrors.Is(errTestSentinel, errTestSentinel)
	if !want {
		t.Fatal("sentinel must match itself")
	}
}

func testJoin(a, b error) error {
	return stderrors.Join(a, b) // want `GID-146: errors\.Join is forbidden\. Fix: use only github\.com/pkg/errors for errors`
}
