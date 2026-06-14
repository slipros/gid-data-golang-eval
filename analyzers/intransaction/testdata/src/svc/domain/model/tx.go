// Canonical transaction model. Eval for GID-175 (naming + negative).
package model

import "context"

// Canonical form — names are correct, no diagnostic.

type InTransactionFunc func(ctx context.Context, fn func(ctx context.Context) error) error

type InTransactionWithReturnFunc[T any] func(ctx context.Context, fn func(ctx context.Context) (T, error)) (T, error)

// NewInTransactionWithReturnFunc wraps InTransactionFunc to return a value.
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

// --- Check 2 (positive): wrong tx-type name in model ---

type RunInTx func(ctx context.Context, fn func(ctx context.Context) error) error // want `GID-175: the transaction type must be named InTransactionFunc / InTransactionWithReturnFunc\. Fix: rename it`

type WithTxResult[T any] func(ctx context.Context, fn func(ctx context.Context) (T, error)) (T, error) // want `GID-175: the transaction type must be named InTransactionFunc / InTransactionWithReturnFunc\. Fix: rename it`

// Edge case in model: a similar but different signature — not flagged.
// Callback with an extra argument.
type NotTx func(ctx context.Context, fn func(ctx context.Context, id int) error) error
