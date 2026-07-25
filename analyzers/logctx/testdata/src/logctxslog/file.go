// Eval for GID-155 on the slog stack: slog has no WithContext — the context
// travels in the *Context variant of the terminal call — and no WithError:
// the error goes in as an attribute.
package logctxslog

import (
	"context"
	"log/slog"
)

type Svc struct {
	logger *slog.Logger
}

// --- Positive cases ---

func (s *Svc) badNoCtx(ctx context.Context) {
	s.logger.Info("start") // want `GID-155: a log call in a function with ctx must carry the context\. Fix: add WithContext\(ctx\) \(logrus\) or call the \*Context variant`
}

func (s *Svc) badErrorNoErr(ctx context.Context) {
	s.logger.ErrorContext(ctx, "failed") // want `GID-155: an Error-level log must carry the error\. Fix: add WithError\(err\) \(logrus\) or pass the error as an attribute`
}

func (s *Svc) badBoth(ctx context.Context) {
	s.logger.Error("failed") // want `GID-155: a log call in a function with ctx must carry the context\. Fix: add WithContext\(ctx\) \(logrus\) or call the \*Context variant` `GID-155: an Error-level log must carry the error\. Fix: add WithError\(err\) \(logrus\) or pass the error as an attribute`
}

// --- Negative cases ---

func (s *Svc) good(ctx context.Context, err error) {
	s.logger.ErrorContext(ctx, "failed", slog.Any("error", err))
}

func (s *Svc) goodInfo(ctx context.Context) {
	s.logger.InfoContext(ctx, "start", slog.String("step", "start"))
}

// Boundary: the enrichment chain of slog is With, not WithField.
func (s *Svc) goodWith(ctx context.Context, err error) {
	s.logger.
		With(slog.String("step", "commit")).
		ErrorContext(ctx, "failed", slog.Any("error", err))
}

// --- Non-applicability: a function without ctx, level not Error ---

func (s *Svc) notApplicable() {
	s.logger.Info("no ctx required")
}
