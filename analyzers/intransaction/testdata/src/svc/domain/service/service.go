// Eval for GID-175: check 3 (anonymous signature) and check 4 (tx-method on a service).
package service

import (
	"context"

	"svc/domain/model"
)

// --- Negative: a field of the named type model.InTransactionFunc — ok ---

type JobService struct {
	tx model.InTransactionFunc
}

func NewJobService(tx model.InTransactionFunc) *JobService {
	return &JobService{tx: tx}
}

// --- Check 3 (positive): anonymous tx-signature in a struct field ---

type BadService struct {
	tx func(ctx context.Context, fn func(ctx context.Context) error) error // want `GID-175: use the named type model.InTransactionFunc\. Fix: replace the anonymous signature`
}

// --- Check 3 (positive): anonymous tx-signature in a constructor parameter ---

func NewBadService(tx func(ctx context.Context, fn func(ctx context.Context) error) error) *BadService { // want `GID-175: use the named type model.InTransactionFunc\. Fix: replace the anonymous signature`
	return &BadService{tx: tx}
}

// --- Check 4 (positive): tx-method on a service ---

func (s *JobService) Transaction(ctx context.Context, fn func(ctx context.Context) error) error { // want `GID-175: a repository/service must not wrap a transaction in a method`
	return s.tx(ctx, fn)
}

// --- Edge case: a method with a similar but different signature (callback without ctx) — not flagged ---

func (s *JobService) NotTransaction(ctx context.Context, fn func() error) error {
	return nil
}
