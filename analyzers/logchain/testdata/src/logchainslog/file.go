// Eval for GID-156 on the slog stack: the enrichment method is With, but the
// per-line requirement for a chain is the same.
package logchainslog

import (
	"context"
	"log/slog"
)

type Svc struct {
	logger *slog.Logger
}

// --- Positive: a chain of two calls on one line ---

func (s *Svc) bad(ctx context.Context) {
	s.logger.With(slog.String("step", "start")).InfoContext(ctx, "start") // want `GID-156: a logger chain must put one call per line, including the first\. Fix: break each call onto a new line`
}

// --- Negative: one call per line ---

func (s *Svc) good(ctx context.Context) {
	s.logger.
		With(slog.String("step", "start")).
		InfoContext(ctx, "start")
}

// --- Non-applicability: a single inline call is allowed ---

func (s *Svc) single(ctx context.Context) {
	s.logger.InfoContext(ctx, "start")
}
