// Eval of GID-176 (part 1): the /dal/repository boundary.
package repository

import (
	"github.com/pkg/errors"

	"boundarysvc/dal/entity"
)

type Repo struct{}

func (r *Repo) call() error { return nil }

func (r *Repo) callRow() (int, error) { return 0, nil }

// --- Positive: pass-through of a non-static error from a call ---

func (r *Repo) badPassThrough() error {
	err := r.call()
	return err // want `GID-176: wrap with errors\.Wrap\. Fix: an error from the app boundary must collect stack and context`
}

func (r *Repo) badPassThroughMulti() (int, error) {
	n, err := r.callRow()
	return n, err // want `GID-176: wrap with errors\.Wrap\. Fix: an error from the app boundary must collect stack and context`
}

// --- Positive: WithStack/WithMessage add no context ---

func (r *Repo) badWithStack() error {
	err := r.call()
	return errors.WithStack(err) // want `GID-176: an error from the app boundary must be wrapped with errors\.Wrap\. Fix: collect stack and context \(WithStack adds no context\)`
}

func (r *Repo) badWithMessage() error {
	err := r.call()
	return errors.WithMessage(err, "ctx") // want `GID-176: an error from the app boundary must be wrapped with errors\.Wrap\. Fix: collect stack and context \(WithMessage adds no context\)`
}

// --- Negative: the error from a call is wrapped with Wrap ---

func (r *Repo) goodWrap() error {
	err := r.call()
	return errors.Wrap(err, "select")
}

// --- Boundary: returning a static error (GID-177 territory, not GID-176) ---

func (r *Repo) goodStatic() error {
	err := r.call()
	if err != nil {
		return entity.ErrNotFound
	}
	return nil
}

// --- Inapplicable: the function does not return an error ---

func (r *Repo) noError() int {
	_ = r.call()
	return 0
}
