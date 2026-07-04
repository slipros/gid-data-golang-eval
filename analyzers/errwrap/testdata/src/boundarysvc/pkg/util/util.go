// Eval of GID-176 (part 1, v2): /pkg/util is not a scoped boundary layer, so an
// interface-method call here is not tracked — but a direct call into a package
// outside the current module (mechanism a) is a boundary in ANY layer,
// including here.
package util

import (
	"strconv"

	"github.com/pkg/errors"
)

type Worker struct{}

func (w *Worker) call() error { return nil }

// --- Non-applicability: not a boundary layer, and call() is a same-module concrete-type call ---

func (w *Worker) passThrough() error {
	err := w.call()
	return err // ok: not a boundary (no client / dal/repository / event in the path, call() is same-module)
}

// --- Positive: a direct external call must be wrapped even outside the scoped boundary layers ---

func (w *Worker) badExternalCall() error {
	_, err := strconv.Atoi("x")
	return err // want `GID-176: an error from an external call must be wrapped with errors\.Wrap\. Fix: collect stack and context`
}

// --- Negative: the external call is wrapped with Wrap ---

func (w *Worker) goodExternalWrap() error {
	_, err := strconv.Atoi("x")
	return errors.Wrap(err, "atoi")
}
