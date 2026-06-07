// Eval GID-214: негатив — composition root internal/app, создание разрешено.
package app

import (
	"github.com/sirupsen/logrus"
)

// New собирает приложение и создаёт корневой логгер — это разрешено.
func New() *logrus.Entry {
	logger := logrus.New()
	std := logrus.StandardLogger()
	_ = std
	return logger.WithField("app", "svc")
}
