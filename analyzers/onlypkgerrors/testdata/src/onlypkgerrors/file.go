// Eval for GID-146 (only-pkg-errors).
package onlypkgerrors

import (
	stderrors "errors"
	"fmt"

	"github.com/pkg/errors"
)

// --- Positive cases: std constructors are caught ---

var ErrStd = stderrors.New("std") // want `GID-146: errors\.New is forbidden\. Fix: use only github\.com/pkg/errors for errors`

func badErrorf(id string) error {
	return fmt.Errorf("job %s failed", id) // want `GID-146: fmt\.Errorf is forbidden\. Fix: use only github\.com/pkg/errors for errors`
}

// Boundary case: errors.Join is also error creation.
func badJoin(a, b error) error {
	return stderrors.Join(a, b) // want `GID-146: errors\.Join is forbidden\. Fix: use only github\.com/pkg/errors for errors`
}

// --- Negative cases: pkg/errors passes ---

var ErrGood = errors.New("good")

func goodWrap(err error) error {
	return errors.Wrap(err, "context")
}

// Boundary case: chain inspection — std Is/As/Unwrap are allowed.
func goodIs(err error) bool {
	return stderrors.Is(err, ErrGood)
}

func goodUnwrap(err error) error {
	return stderrors.Unwrap(err)
}

// --- Not applicable: fmt for strings — not errors ---

func notApplicable(id string) string {
	return fmt.Sprintf("job %s", id)
}
