// Eval of GID-248: errors.WithStack of an error that already carries a stack
// layers a second one. An error counts as already stacked when every assignment
// to it is a call inside the same module (a method of this service, an
// interface call outside the boundary layers, a pkg/errors constructor).
package service

import (
	"strconv"

	"github.com/pkg/errors"

	"stacksvc/domain/model"
)

// Repo is an injected same-module dependency: /domain/service is not a boundary
// layer, so the implementation wraps at its own origin (GID-176) and the error
// arrives here already stacked.
type Repo interface {
	Get(id string) error
}

type Service struct {
	repo Repo
}

func (s *Service) call() error { return nil }

func (s *Service) generateRawSecret() (string, error) { return "", nil }

// --- Positive: WithStack of an error from a same-module method call ---

func (s *Service) badSameModuleCall() (string, error) {
	raw, err := s.generateRawSecret()
	if err != nil {
		return "", errors.WithStack(err) // want `GID-248: errors\.WithStack of an error that already carries a stack layers a second one\. Fix: return the error as is \(return err\)`
	}

	return raw, nil
}

// --- Positive: WithStack of an interface-call error outside the boundary layers ---

func (s *Service) badInterfaceCall(id string) error {
	err := s.repo.Get(id)
	if err != nil {
		return errors.WithStack(err) // want `GID-248: errors\.WithStack of an error that already carries a stack layers a second one\. Fix: return the error as is \(return err\)`
	}

	return nil
}

// --- Positive: WithStack of a pkg/errors constructor result ---

func (s *Service) badPkgErrorsCtor() error {
	err := errors.New("boom")

	return errors.WithStack(err) // want `GID-248: errors\.WithStack of an error that already carries a stack layers a second one\. Fix: return the error as is \(return err\)`
}

// --- Boundary: a same-module call used directly as the argument, no variable in between ---

func (s *Service) badDirectCallArg() error {
	return errors.WithStack(s.call()) // want `GID-248: errors\.WithStack of an error that already carries a stack layers a second one\. Fix: return the error as is \(return err\)`
}

// --- Negative: the error was replaced before WithStack ---

func (s *Service) goodConverted() error {
	err := s.call()
	if err != nil {
		err = model.ErrNotFound
		return errors.WithStack(err)
	}

	return nil
}

// --- Negative: WithStack of a static error — that is what GID-177 demands ---

func (s *Service) goodStaticVar() error {
	return errors.WithStack(model.ErrNotFound)
}

func (s *Service) goodStaticLiteral() error {
	return errors.WithStack(&model.BigError{Code: 1})
}

// --- Negative: WithStack of a function parameter — the origin is unknown ---

func mapErr(err error) error {
	return errors.WithStack(err)
}

// --- Negative: the error is returned as is — the canonical fix ---

func (s *Service) goodPassThrough() error {
	err := s.call()
	if err != nil {
		return err
	}

	return nil
}

// --- Boundary: one of the assignments has an unknown source ---

func (s *Service) goodMixedSource(errCh chan error, cond bool) error {
	err := s.call()
	if cond {
		err = <-errCh
	}

	return errors.WithStack(err)
}

// --- Non-applicability: an external call — GID-176 demands errors.Wrap there ---

func (s *Service) externalCall() error {
	_, err := strconv.Atoi("x")

	return errors.WithStack(err)
}

// --- Non-applicability: settings.exclude exempts a specific method ---

func (s *Service) excludedMethod() error {
	err := s.call()

	return errors.WithStack(err)
}
