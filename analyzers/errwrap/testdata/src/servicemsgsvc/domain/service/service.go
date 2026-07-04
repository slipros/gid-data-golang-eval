// Eval of GID-237: errors.WithMessage/WithMessagef is banned in /domain/service —
// a service converts the error and wraps it with errors.WithStack; adding a
// message to an incoming error belongs to /domain/usecase.
package service

import "github.com/pkg/errors"

type Service struct{}

func (s *Service) call() error { return nil }

// --- Positive: errors.WithMessage in a service ---

func (s *Service) badWithMessage() error {
	err := s.call()
	return errors.WithMessage(err, "ctx") // want `GID-237: errors\.WithMessage is not used in a service — convert the error and wrap with errors\.WithStack; WithMessage belongs to usecase`
}

// --- Boundary: errors.WithMessagef — the formatted variant is banned too ---

func (s *Service) badWithMessagef() error {
	err := s.call()
	return errors.WithMessagef(err, "ctx %d", 1) // want `GID-237: errors\.WithMessage is not used in a service — convert the error and wrap with errors\.WithStack; WithMessage belongs to usecase`
}

// --- Negative: errors.WithStack / errors.Wrap are fine in a service ---

func (s *Service) goodWithStack() error {
	err := s.call()
	return errors.WithStack(err)
}

func (s *Service) goodWrap() error {
	err := s.call()
	return errors.Wrap(err, "ctx")
}

// --- Non-applicability: settings.exclude exempts a specific method ---

func (s *Service) excludedMethod() error {
	err := s.call()
	return errors.WithMessage(err, "ctx")
}
