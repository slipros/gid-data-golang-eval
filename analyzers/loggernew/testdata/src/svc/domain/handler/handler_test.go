// Eval GID-214: not applicable — _test.go, a logger in tests is allowed.
package handler

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestSomething(t *testing.T) {
	// In tests creating a logger is fine — _test.go is skipped.
	logger := logrus.New()
	_ = logger
}
