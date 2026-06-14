// Eval GID-214: negative — composition root internal/app, creation is allowed.
package app

import (
	"github.com/sirupsen/logrus"
)

// New assembles the application and creates the root logger — this is allowed.
func New() *logrus.Entry {
	logger := logrus.New()
	std := logrus.StandardLogger()
	_ = std
	return logger.WithField("app", "svc")
}
