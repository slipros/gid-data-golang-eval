// Eval GID-176 (часть 2): /domain/** — Wrap нестатичной ошибки запрещён.
package service

import (
	"github.com/pkg/errors"

	"domainsvc/domain/model"
)

type Service struct{}

func (s *Service) call() error { return nil }

// --- Позитив: Wrap пришедшей нестатичной ошибки ---

func (s *Service) badWrap() error {
	err := s.call()
	return errors.Wrap(err, "ctx") // want `GID-176: the stack is already collected at the boundary\. Fix: use errors\.WithMessage instead of errors\.Wrap for an incoming error`
}

func (s *Service) badWrapParam(err error) error {
	return errors.Wrap(err, "ctx") // want `GID-176: the stack is already collected at the boundary\. Fix: use errors\.WithMessage instead of errors\.Wrap for an incoming error`
}

// --- Негатив: WithMessage для пришедшей ошибки ---

func (s *Service) goodWithMessage() error {
	err := s.call()
	return errors.WithMessage(err, "ctx")
}

// --- Граничный: Wrap статичной ошибки из model — разрешено ---

func (s *Service) goodWrapStatic() error {
	return errors.Wrap(model.ErrSnapshotNotFound, "ctx")
}

// --- Неприменимость: не Wrap (WithStack пришедшей ошибки) — здесь GID-176-часть-2 молчит ---

func (s *Service) notWrap() error {
	err := s.call()
	return errors.WithStack(err)
}
