// Eval for GID-175: check 4 (tx-method on a repository).
package repository

import "context"

type JobRepository struct{}

// --- Check 4 (positive): tx-method on a repository (any name) ---

func (r *JobRepository) InTx(ctx context.Context, fn func(ctx context.Context) error) error { // want `GID-175: a repository/service must not wrap a transaction in a method`
	return fn(ctx)
}

// --- Edge case: a method with a similar signature, callback returns non-error — not flagged ---

func (r *JobRepository) NotInTx(ctx context.Context, fn func(ctx context.Context) (int, error)) error {
	return nil
}

// --- Negative: an ordinary repository method ---

func (r *JobRepository) Job(ctx context.Context, id string) error {
	return nil
}
