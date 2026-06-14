// Eval GID-214: positive — logrus.StandardLogger() in /dal/repository.
package repository

import (
	"github.com/sirupsen/logrus"
)

type Repo struct {
	logger *logrus.Logger
}

// Positive case: StandardLogger() is also forbidden outside the composition root.
func New() *Repo {
	return &Repo{
		logger: logrus.StandardLogger(), // want `GID-214: logrus.StandardLogger\(\) may be called only in the composition root \(main, internal/app\)\. Fix: pass a ready \*logrus\.Entry through the constructor`
	}
}
