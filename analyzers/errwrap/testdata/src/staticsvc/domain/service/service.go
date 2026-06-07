// Eval GID-177: статичные ошибки при возврате оборачиваются WithStack.
package service

import (
	"github.com/pkg/errors"

	"staticsvc/domain/model"
	"staticsvc/pkg/gderror"
)

type Service struct{}

// --- Позитив: package-level error-var без обёртки ---

func (s *Service) badVar() error {
	return model.ErrSnapshotNotFound // want `GID-177: a static error is returned without a stack\. Fix: wrap with errors\.WithStack \(or errors\.Wrap if you need context\)`
}

// --- Позитив: адрес именованного error-типа без обёртки ---

func (s *Service) badPtrLit() error {
	return &model.BigError{Code: 1} // want `GID-177: a static error is returned without a stack\. Fix: wrap with errors\.WithStack \(or errors\.Wrap if you need context\)`
}

// --- Позитив: композит-литерал именованного error-типа без обёртки ---

func (s *Service) badValueLit() error {
	return model.BigError{Code: 2} // want `GID-177: a static error is returned without a stack\. Fix: wrap with errors\.WithStack \(or errors\.Wrap if you need context\)`
}

// --- Негатив: WithStack / Wrap статичной ошибки ---

func (s *Service) goodWithStack() error {
	return errors.WithStack(model.ErrSnapshotNotFound)
}

func (s *Service) goodWrap() error {
	return errors.Wrap(model.ErrSnapshotNotFound, "ctx")
}

// --- Граничный: возврат пришедшей нестатичной ошибки — не GID-177 ---

func (s *Service) goodPassThrough(err error) error {
	return err
}

// --- Неприменимость: конструктор-исключение сам собирает стек (settings.exclude) ---

func (s *Service) goodExcludedCtor() error {
	return gderror.NewUnhandledValueError("x")
}

// --- Неприменимость: функция без error ---

func (s *Service) noError() int {
	return 0
}
