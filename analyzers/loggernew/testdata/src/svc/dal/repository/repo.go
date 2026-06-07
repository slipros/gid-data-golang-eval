// Eval GID-214: позитив — logrus.StandardLogger() в /dal/repository.
package repository

import (
	"github.com/sirupsen/logrus"
)

type Repo struct {
	logger *logrus.Logger
}

// Позитивный кейс: StandardLogger() тоже запрещён вне composition root.
func New() *Repo {
	return &Repo{
		logger: logrus.StandardLogger(), // want `GID-214: logrus.StandardLogger\(\) may be called only in the composition root \(main, internal/app\)\. Fix: pass a ready \*logrus\.Entry through the constructor`
	}
}
