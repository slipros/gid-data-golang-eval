// Eval for GID-175: check 3 (anonymous signature) in usecase.
package usecase

import (
	"context"

	"svc/domain/model"
)

// --- Negative: the named type model.InTransactionFunc in a field — ok ---

type JobUsecase struct {
	tx model.InTransactionFunc
}

// --- Check 3 (positive): anonymous generic tx-signature in a function parameter ---

func WithTx(run func(ctx context.Context, fn func(ctx context.Context) (string, error)) (string, error)) { // want `GID-175: use the named type model.InTransactionFunc\. Fix: replace the anonymous signature`
	_ = run
}

// --- Edge case: callback returns non-error — not flagged ---

func NotTx(run func(ctx context.Context, fn func(ctx context.Context) int) int) {
	_ = run
}
