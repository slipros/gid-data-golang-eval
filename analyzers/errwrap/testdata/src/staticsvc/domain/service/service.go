// Eval of GID-177: static errors are wrapped with WithStack on return.
package service

import (
	"github.com/pkg/errors"

	"staticsvc/domain/model"
	"staticsvc/pkg/gderror"
)

type Service struct{}

// --- Positive: a package-level error var without a wrapper ---

func (s *Service) badVar() error {
	return model.ErrSnapshotNotFound // want `GID-177: a static error is returned without a stack\. Fix: wrap with errors\.WithStack \(or errors\.Wrap if you need context\)`
}

// --- Positive: the address of a named error type without a wrapper ---

func (s *Service) badPtrLit() error {
	return &model.BigError{Code: 1} // want `GID-177: a static error is returned without a stack\. Fix: wrap with errors\.WithStack \(or errors\.Wrap if you need context\)`
}

// --- Positive: a composite literal of a named error type without a wrapper ---

func (s *Service) badValueLit() error {
	return model.BigError{Code: 2} // want `GID-177: a static error is returned without a stack\. Fix: wrap with errors\.WithStack \(or errors\.Wrap if you need context\)`
}

// --- Negative: WithStack / Wrap of a static error ---

func (s *Service) goodWithStack() error {
	return errors.WithStack(model.ErrSnapshotNotFound)
}

func (s *Service) goodWrap() error {
	return errors.Wrap(model.ErrSnapshotNotFound, "ctx")
}

// --- Boundary: returning an incoming non-static error — not GID-177 ---

func (s *Service) goodPassThrough(err error) error {
	return err
}

// --- Inapplicable: an excluded constructor collects the stack itself (settings.exclude) ---

func (s *Service) goodExcludedCtor() error {
	return gderror.NewUnhandledValueError("x")
}

// --- Inapplicable: a function without an error ---

func (s *Service) noError() int {
	return 0
}
