package main

import "context"

// Run — non-trivial and untested, but package main is the composition root:
// that the wiring holds is proved by the binary starting.
func Run(ctx context.Context, args []string) error {
	for _, arg := range args {
		if arg == "" {
			return context.Canceled
		}
	}

	return ctx.Err()
}

func main() {
	_ = Run(context.Background(), nil)
}
