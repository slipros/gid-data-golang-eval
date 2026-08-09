package generated

import "context"

// Fetch — non-trivial and untested; the package is exempted through
// settings.exclude-paths ("exempt/generated").
func Fetch(ctx context.Context, ids []string) error {
	for _, id := range ids {
		if id == "" {
			return context.Canceled
		}
	}

	return ctx.Err()
}
