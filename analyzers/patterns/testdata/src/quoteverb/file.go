package quoteverb

import (
	"fmt"

	"github.com/pkg/errors"
)

// Позитив: ручное экранирование кавычек вокруг %s/%v.
func bad(name string) string {
	s := fmt.Sprintf("\"%s\"", name)  // want `GID-007: do not escape quotes around %s/%v by hand\. Fix: use %q instead of \\"%s\\"\.`
	_ = errors.Errorf("\"%v\"", name) // want `GID-007: do not escape quotes`
	return s
}

// Негатив: %q вместо ручного экранирования.
func good(name string) string {
	return fmt.Sprintf("%q", name)
}

// Неприменимость: формат без обёрнутых кавычек.
func boundary(n int) string {
	return fmt.Sprintf("count=%d", n)
}
