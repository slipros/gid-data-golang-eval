// Eval of GID-237: errors.WithMessage/WithMessagef is banned in /domain/service —
// adding a message to an incoming error belongs to /domain/usecase. An incoming
// error is returned as is (its stack is already collected upstream);
// errors.WithStack is for an error the service converted itself.
package service

import "github.com/pkg/errors"

// ErrConverted — the service's own model error (no stack of its own).
var ErrConverted = errors.New("converted")

type Service struct{}

func (s *Service) call() error { return nil }

// --- Positive: errors.WithMessage in a service ---

func (s *Service) badWithMessage() error {
	err := s.call()
	return errors.WithMessage(err, "ctx") // want `GID-237: errors\.WithMessage is not used in a service\. Fix: return the incoming error as is`
}

// --- Boundary: errors.WithMessagef — the formatted variant is banned too ---

func (s *Service) badWithMessagef() error {
	err := s.call()
	return errors.WithMessagef(err, "ctx %d", 1) // want `GID-237: errors\.WithMessage is not used in a service\. Fix: return the incoming error as is`
}

// --- Negative: an incoming error passed through as is — the canonical fix ---

func (s *Service) goodPassThrough() error {
	err := s.call()
	if err != nil {
		return err
	}

	return nil
}

// --- Negative: errors.WithStack of a converted error / errors.Wrap ---

func (s *Service) goodWithStackConverted() error {
	err := s.call()
	if err != nil {
		err = ErrConverted
		return errors.WithStack(err)
	}

	return nil
}

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
