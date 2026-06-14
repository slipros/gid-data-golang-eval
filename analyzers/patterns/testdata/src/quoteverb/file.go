package quoteverb

import (
	"fmt"

	"github.com/pkg/errors"
)

// Positive: manually escaped quotes around %s/%v.
func bad(name string) string {
	s := fmt.Sprintf("\"%s\"", name)  // want `GID-007: do not escape quotes around %s/%v by hand\. Fix: use %q instead of \\"%s\\"\.`
	_ = errors.Errorf("\"%v\"", name) // want `GID-007: do not escape quotes`
	return s
}

// Negative: %q instead of manual escaping.
func good(name string) string {
	return fmt.Sprintf("%q", name)
}

// Not applicable: a format without wrapped quotes.
func boundary(n int) string {
	return fmt.Sprintf("count=%d", n)
}
