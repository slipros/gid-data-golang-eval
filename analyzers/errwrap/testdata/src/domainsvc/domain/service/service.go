// Eval of GID-176 (part 2): /domain/** — Wrap of a non-static error is forbidden.
package service

import (
	"github.com/pkg/errors"

	"domainsvc/domain/model"
)

type Service struct{}

func (s *Service) call() error { return nil }

// --- Positive: Wrap of an incoming non-static error ---

func (s *Service) badWrap() error {
	err := s.call()
	return errors.Wrap(err, "ctx") // want `GID-176: the stack is already collected at the boundary\. Fix: use errors\.WithMessage instead of errors\.Wrap for an incoming error`
}

func (s *Service) badWrapParam(err error) error {
	return errors.Wrap(err, "ctx") // want `GID-176: the stack is already collected at the boundary\. Fix: use errors\.WithMessage instead of errors\.Wrap for an incoming error`
}

// --- Negative: WithMessage for an incoming error ---

func (s *Service) goodWithMessage() error {
	err := s.call()
	return errors.WithMessage(err, "ctx")
}

// --- Boundary: Wrap of a static error from model — allowed ---

func (s *Service) goodWrapStatic() error {
	return errors.Wrap(model.ErrSnapshotNotFound, "ctx")
}

// --- Inapplicable: not Wrap (WithStack of an incoming error) — GID-176 part 2 stays silent here ---

func (s *Service) notWrap() error {
	err := s.call()
	return errors.WithStack(err)
}
