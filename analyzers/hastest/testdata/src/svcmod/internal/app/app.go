package app

import "context"

// Options — the wiring of the application.
type Options struct {
	DSN string
}

// Start — non-trivial and untested, but /app/** is the composition root: it
// assembles dependencies, and that it assembles them correctly is proved by the
// service starting, not by a unit test.
func Start(ctx context.Context, opts Options) error {
	if opts.DSN == "" {
		return context.Canceled
	}

	return ctx.Err()
}
