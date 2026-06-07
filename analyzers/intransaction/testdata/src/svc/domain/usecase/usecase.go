// Eval для GID-175: проверка 3 (анонимная сигнатура) в usecase.
package usecase

import (
	"context"

	"svc/domain/model"
)

// --- Негатив: именованный тип model.InTransactionFunc в поле — ок ---

type JobUsecase struct {
	tx model.InTransactionFunc
}

// --- Проверка 3 (позитив): анонимная generic-tx-сигнатура в параметре функции ---

func WithTx(run func(ctx context.Context, fn func(ctx context.Context) (string, error)) (string, error)) { // want `GID-175: use the named type model.InTransactionFunc\. Fix: replace the anonymous signature`
	_ = run
}

// --- Граничный кейс: callback возвращает не error — не флагуем ---

func NotTx(run func(ctx context.Context, fn func(ctx context.Context) int) int) {
	_ = run
}
