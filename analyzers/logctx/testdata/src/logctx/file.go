// Eval for GID-155 (WithContext / WithError).
package logctx

import (
	"context"

	"github.com/sirupsen/logrus"
)

type Svc struct {
	logger *logrus.Entry
}

// --- Positive cases ---

func (s *Svc) badNoCtx(ctx context.Context) {
	s.logger.Info("start") // want `GID-155: a log call in a function with ctx must include WithContext\(ctx\)\. Fix: add WithContext\(ctx\)`
}

func (s *Svc) badErrorNoErr(ctx context.Context, err error) {
	s.logger.
		WithContext(ctx).
		Error("failed") // want `GID-155: an Error-level log must include WithError\(err\)\. Fix: add WithError\(err\)`
}

// Boundary case: both violations in a single call.
func (s *Svc) badBoth(ctx context.Context) {
	s.logger.Error("failed") // want `GID-155: a log call in a function with ctx must include WithContext\(ctx\)\. Fix: add WithContext\(ctx\)` `GID-155: an Error-level log must include WithError\(err\)\. Fix: add WithError\(err\)`
}

// Boundary case: the outer function has ctx, but the log is inside a closure
// without ctx — no requirement is imposed on the closure.
func (s *Svc) closure(ctx context.Context) func() {
	return func() {
		s.logger.Info("tick")
	}
}

// --- Negative cases ---

func (s *Svc) good(ctx context.Context, err error) {
	s.logger.
		WithContext(ctx).
		WithError(err).
		Error("failed")
}

func (s *Svc) goodInfo(ctx context.Context) {
	s.logger.
		WithContext(ctx).
		WithField("step", "start").
		Info("start")
}

// --- Non-applicability: a function without ctx, level not Error ---

func (s *Svc) notApplicable() {
	s.logger.Info("no ctx required")
}
