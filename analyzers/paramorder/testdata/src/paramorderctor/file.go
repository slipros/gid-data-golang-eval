// Eval of GID-252: in a constructor the known parameters keep the order
// ctx, opts, logger, metrics and all of them precede the entity's own
// dependencies. Any of them may be absent.
package paramorderctor

import (
	"context"
	"log/slog"
)

type InterceptorOptions struct {
	Window int
}

// InterceptorMetrics — the metrics struct of the entity (GID-174 shape).
type InterceptorMetrics struct {
	Hits int
}

// BanChecker — the entity's own dependency.
type BanChecker interface {
	Banned(id string) bool
}

type Interceptor struct{}

// --- Positive: the logger sits after a dependency ---

func NewInterceptor(opts InterceptorOptions, checker BanChecker, logger *slog.Logger) *Interceptor { // want `GID-252: the logger comes after the entity's own dependencies\. Fix: order the parameters ctx, opts, logger, metrics, then the rest`
	return &Interceptor{}
}

// --- Positive: metrics after a dependency ---

func NewCollector(checker BanChecker, metrics *InterceptorMetrics) *Interceptor { // want `GID-252: metrics come after the entity's own dependencies\. Fix: order the parameters ctx, opts, logger, metrics, then the rest`
	return &Interceptor{}
}

// --- Positive: metrics before the logger ---

func NewSwapped(metrics *InterceptorMetrics, logger *slog.Logger) *Interceptor { // want `GID-252: metrics come before the logger\. Fix: order the parameters ctx, opts, logger, metrics, then the rest`
	return &Interceptor{}
}

// --- Negative: the full canonical order ---

func NewFull(ctx context.Context, opts InterceptorOptions, logger *slog.Logger, metrics *InterceptorMetrics, checker BanChecker) *Interceptor {
	return &Interceptor{}
}

// --- Boundary: parts of the prefix are absent, the present ones are ordered ---

func NewPartial(logger *slog.Logger, checker BanChecker) *Interceptor {
	return &Interceptor{}
}

func NewDepsOnly(checker BanChecker) *Interceptor {
	return &Interceptor{}
}

// --- Non-applicability: an ordinary function is not a constructor ---

func enrich(checker BanChecker, logger *slog.Logger) *Interceptor {
	return &Interceptor{}
}
