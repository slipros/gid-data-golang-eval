// Eval of GID-176 (part 2, v2): /domain/** — Wrap of a same-module non-static
// error is forbidden (its stack, if any, was already collected upstream), but
// Wrap of an error from a direct external call (mechanism a) is required —
// the domain may be the first place that reaches out to an external
// dependency, e.g. a DB connection or a stdlib call.
package service

import (
	"strconv"

	"github.com/pkg/errors"

	"domainsvc/domain/model"
)

type Service struct{}

func (s *Service) call() error { return nil }

// --- Positive: Wrap of an incoming same-module non-static error ---

func (s *Service) badWrap() error {
	err := s.call()
	return errors.Wrap(err, "ctx") // want `GID-176: the stack is already collected upstream for a same-module error\. Fix: use errors\.WithMessage instead of errors\.Wrap for an incoming error`
}

func (s *Service) badWrapParam(err error) error {
	return errors.Wrap(err, "ctx") // want `GID-176: the stack is already collected upstream for a same-module error\. Fix: use errors\.WithMessage instead of errors\.Wrap for an incoming error`
}

// --- Negative: WithMessage for a same-module incoming error ---

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

// --- Positive (v2): a direct external call in /domain/service must be wrapped ---

func (s *Service) badExternalCall() error {
	_, err := strconv.Atoi("x")
	return err // want `GID-176: an error from an external call must be wrapped with errors\.Wrap\. Fix: collect stack and context`
}

// --- Negative (v2): the external call is correctly wrapped — Wrap is required, not forbidden, in domain too ---

func (s *Service) goodExternalWrap() error {
	_, err := strconv.Atoi("x")
	return errors.Wrap(err, "parse")
}
