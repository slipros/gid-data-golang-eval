// Eval для GID-146 (only-pkg-errors).
package onlypkgerrors

import (
	stderrors "errors"
	"fmt"

	"github.com/pkg/errors"
)

// --- Позитивные кейсы: std-конструкторы пойманы ---

var ErrStd = stderrors.New("std") // want `GID-146: errors\.New is forbidden\. Fix: use only github\.com/pkg/errors for errors`

func badErrorf(id string) error {
	return fmt.Errorf("job %s failed", id) // want `GID-146: fmt\.Errorf is forbidden\. Fix: use only github\.com/pkg/errors for errors`
}

// Граничный кейс: errors.Join — тоже создание ошибки.
func badJoin(a, b error) error {
	return stderrors.Join(a, b) // want `GID-146: errors\.Join is forbidden\. Fix: use only github\.com/pkg/errors for errors`
}

// --- Негативные кейсы: pkg/errors проходит ---

var ErrGood = errors.New("good")

func goodWrap(err error) error {
	return errors.Wrap(err, "context")
}

// Граничный кейс: проверка цепочки — std Is/As/Unwrap разрешены.
func goodIs(err error) bool {
	return stderrors.Is(err, ErrGood)
}

func goodUnwrap(err error) error {
	return stderrors.Unwrap(err)
}

// --- Неприменимость: fmt для строк — не ошибки ---

func notApplicable(id string) string {
	return fmt.Sprintf("job %s", id)
}
