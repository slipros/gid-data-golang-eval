// Eval for GID-175: check 1 applies everywhere except /domain/model.
// helper is not part of the service/usecase/repository layers, but check 1
// still flags the tx-type declaration here.
package helper

import "context"

// --- Check 1 (positive): a named tx-type declared outside /domain/model ---

type Tx func(ctx context.Context, fn func(ctx context.Context) error) error // want `GID-175: the transaction type must live in /domain/model \(InTransactionFunc\)\. Fix: move it there`

// --- Check 1 (positive): generic variant outside model ---

type TxR[T any] func(ctx context.Context, fn func(ctx context.Context) (T, error)) (T, error) // want `GID-175: the transaction type must live in /domain/model \(InTransactionFunc\)\. Fix: move it there`

// Edge case: a similar but different signature (without ctx) — not flagged.
type NotTx func(fn func(ctx context.Context) error) error
