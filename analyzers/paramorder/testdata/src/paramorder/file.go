// Eval for GID-110/113/153 (param order).
package paramorder

import (
	"context"

	"github.com/sirupsen/logrus"
)

type HelloOptions struct {
	Retries int
}

type Hello struct{}

// --- Positive cases ---

func (h *Hello) BadCtx(id int, ctx context.Context) error { // want `GID-110: context\.Context must be the first parameter\. Fix: move ctx first`
	return nil
}

func (h *Hello) BadOpts(ctx context.Context, id int, opts *HelloOptions) error { // want `GID-113: opts must come right after ctx, not last\. Fix: move opts after ctx`
	return nil
}

// Boundary case: without ctx opts still goes first.
func NewBad(logger *logrus.Entry, opts *HelloOptions) *Hello { // want `GID-113: opts must come right after ctx, not last\. Fix: move opts after ctx` `GID-153: logger must come after the entity opts\. Fix: move logger after opts`
	return &Hello{}
}

// --- Negative cases ---

func (h *Hello) Good(ctx context.Context, opts *HelloOptions, id int) error {
	return nil
}

func NewGood(opts *HelloOptions, logger *logrus.Entry) *Hello {
	return &Hello{}
}

func GoodNoCtx(opts *HelloOptions, size int64) int64 {
	return size
}

// Boundary case: logger without opts — the order is not regulated by this rule.
func NewNoOpts(logger *logrus.Entry, retries int) *Hello {
	return &Hello{}
}

// --- Not applicable: neither ctx, nor opts, nor logger ---

func Plain(a, b int) int { return a + b }
