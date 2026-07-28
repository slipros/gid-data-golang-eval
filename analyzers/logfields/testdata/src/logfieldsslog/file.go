// Eval for GID-253: the slog stack is out of scope — With takes variadic
// pairs, so there is no single-field method to repeat.
package logfieldsslog

import (
	"context"
	"log/slog"
)

type Svc struct {
	logger *slog.Logger
}

func (s *Svc) chained(ctx context.Context, offset int64, level int) {
	s.logger.
		With("offset", offset).
		With("fallback_level", level).
		InfoContext(ctx, "consumed")
}

func (s *Svc) batched(ctx context.Context, offset int64, level int) {
	s.logger.
		With("offset", offset, "fallback_level", level).
		InfoContext(ctx, "consumed")
}
