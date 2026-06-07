// Каноническая модель транзакций. Eval для GID-175 (нейминг + негатив).
package model

import "context"

// Каноническая форма — имена верные, диагностики нет.

type InTransactionFunc func(ctx context.Context, fn func(ctx context.Context) error) error

type InTransactionWithReturnFunc[T any] func(ctx context.Context, fn func(ctx context.Context) (T, error)) (T, error)

// NewInTransactionWithReturnFunc оборачивает InTransactionFunc для возврата значения.
func NewInTransactionWithReturnFunc[T any](tx InTransactionFunc) InTransactionWithReturnFunc[T] {
	return func(ctx context.Context, fn func(ctx context.Context) (T, error)) (T, error) {
		var result T
		err := tx(ctx, func(ctx context.Context) error {
			var innerErr error
			result, innerErr = fn(ctx)
			return innerErr
		})
		return result, err
	}
}

// --- Проверка 2 (позитив): неверное имя tx-типа в model ---

type RunInTx func(ctx context.Context, fn func(ctx context.Context) error) error // want `GID-175: the transaction type must be named InTransactionFunc / InTransactionWithReturnFunc\. Fix: rename it`

type WithTxResult[T any] func(ctx context.Context, fn func(ctx context.Context) (T, error)) (T, error) // want `GID-175: the transaction type must be named InTransactionFunc / InTransactionWithReturnFunc\. Fix: rename it`

// Граничный кейс в model: похожая, но другая сигнатура — не флагуем.
// Callback с дополнительным аргументом.
type NotTx func(ctx context.Context, fn func(ctx context.Context, id int) error) error
