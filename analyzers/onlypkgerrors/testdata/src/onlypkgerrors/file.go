// Eval for GID-146 (only-pkg-errors).
package onlypkgerrors

import (
	stderrors "errors" // want `GID-146: the std errors package is forbidden, github\.com/pkg/errors re-exports Is/As/Unwrap\. Fix: import "github\.com/pkg/errors" alone and call errors\.Is\(err, ErrNoResult\)`
	"fmt"

	"github.com/pkg/errors"
)

// --- Positive: the std errors package is reported on its import, under any alias ---

// The calls below are the same defect as the import, so they carry no diagnostic
// of their own: dropping the import is the single fix for the whole file.

var ErrStd = stderrors.New("std")

func badJoin(a, b error) error {
	return stderrors.Join(a, b)
}

func badIs(err error) bool {
	return stderrors.Is(err, ErrStd)
}

// Positive: fmt.Errorf is reported on the call — the fmt import stays legitimate.

func badErrorf(id string) error {
	return fmt.Errorf("job %s failed", id) // want `GID-146: fmt\.Errorf is forbidden\. Fix: use only github\.com/pkg/errors for errors`
}

// --- Negative: pkg/errors passes ---

var ErrGood = errors.New("good")

func goodWrap(err error) error {
	return errors.Wrap(err, "context")
}

// --- Not applicable: fmt for strings — not errors ---

func notApplicable(id string) string {
	return fmt.Sprintf("job %s", id)
}
