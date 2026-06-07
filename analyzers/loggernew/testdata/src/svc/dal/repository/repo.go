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
		logger: logrus.StandardLogger(), // want `GID-214: logrus.StandardLogger\(\) вызывается только в composition root \(main, internal/app\) — пробрасывай готовый \*logrus\.Entry через конструктор`
	}
}
