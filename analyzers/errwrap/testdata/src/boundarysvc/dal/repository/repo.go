// Eval GID-176 (часть 1): граница /dal/repository.
package repository

import (
	"github.com/pkg/errors"

	"boundarysvc/dal/entity"
)

type Repo struct{}

func (r *Repo) call() error { return nil }

func (r *Repo) callRow() (int, error) { return 0, nil }

// --- Позитив: pass-through нестатичной ошибки из вызова ---

func (r *Repo) badPassThrough() error {
	err := r.call()
	return err // want `GID-176: оберните errors\.Wrap — ошибка с границы приложения должна собрать стек и контекст`
}

func (r *Repo) badPassThroughMulti() (int, error) {
	n, err := r.callRow()
	return n, err // want `GID-176: оберните errors\.Wrap — ошибка с границы приложения должна собрать стек и контекст`
}

// --- Позитив: WithStack/WithMessage не добавляют контекст ---

func (r *Repo) badWithStack() error {
	err := r.call()
	return errors.WithStack(err) // want `GID-176: ошибка с границы приложения оборачивается errors\.Wrap — собрать стек и контекст \(WithStack контекста не добавляет\)`
}

func (r *Repo) badWithMessage() error {
	err := r.call()
	return errors.WithMessage(err, "ctx") // want `GID-176: ошибка с границы приложения оборачивается errors\.Wrap — собрать стек и контекст \(WithMessage контекста не добавляет\)`
}

// --- Негатив: ошибка из вызова обёрнута Wrap ---

func (r *Repo) goodWrap() error {
	err := r.call()
	return errors.Wrap(err, "select")
}

// --- Граничный: возврат статичной ошибки (зона GID-177, не GID-176) ---

func (r *Repo) goodStatic() error {
	err := r.call()
	if err != nil {
		return entity.ErrNotFound
	}
	return nil
}

// --- Неприменимость: функция не возвращает error ---

func (r *Repo) noError() int {
	_ = r.call()
	return 0
}
