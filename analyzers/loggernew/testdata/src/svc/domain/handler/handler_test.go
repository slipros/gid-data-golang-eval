// Eval GID-214: неприменимость — _test.go, логгер в тестах разрешён.
package handler

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestSomething(t *testing.T) {
	// В тестах создавать логгер можно — _test.go пропускается.
	logger := logrus.New()
	_ = logger
}
