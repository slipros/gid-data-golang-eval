// Boundary: the file imports the std package alone, without pkg/errors next to
// it, and only inspects the chain. Before 2026-08-06 this was the shape the rule
// blessed ("pkg/errors has no Is") — it is now the shape the rule is about.
package onlypkgerrors

import "errors" // want `GID-146: the std errors package is forbidden, github\.com/pkg/errors re-exports Is/As/Unwrap\. Fix: import "github\.com/pkg/errors" alone and call errors\.Is\(err, ErrNoResult\)`

func stdOnlyIs(err, target error) bool {
	return errors.Is(err, target)
}

func stdOnlyAs(err error, target any) bool {
	return errors.As(err, target)
}

// Boundary: a local function named Errorf is not fmt.Errorf — the callee is
// resolved through typeutil.Callee, not by name.

func Errorf(format string, args ...any) error { return nil }

func localErrorf(id string) error {
	return Errorf("job %s failed", id)
}
