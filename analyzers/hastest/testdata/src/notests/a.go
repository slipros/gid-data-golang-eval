// Package notests has no test file at all, and no pass can tell that apart
// from a run that withheld the tests (run.tests: false): the rule stays silent
// here by design.
package notests

import "context"

// Build — non-trivial logic in a package with no test file at all.
func Build(items []string) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		if item != "" {
			out = append(out, item)
		}
	}

	return out
}

// Check — a second candidate; the package still gets ONE diagnostic.
func Check(ctx context.Context, id string) error {
	if id == "" {
		return context.Canceled
	}

	return ctx.Err()
}

// Name — a trivial getter: not counted among the candidates.
func Name() string {
	return "notests"
}
