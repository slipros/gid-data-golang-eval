// Eval GID-214: позитив — logrus.New() в /domain/service.
package service

import (
	"github.com/sirupsen/logrus"
)

type Svc struct {
	logger *logrus.Entry
}

// Позитивный кейс: создание логгера в сервисе запрещено.
func New() *Svc {
	l := logrus.New() // want `GID-214: logrus.New\(\) may be called only in the composition root \(main, internal/app\)\. Fix: pass a ready \*logrus\.Entry through the constructor`
	_ = l
	return &Svc{}
}

// Негатив: WithField на готовом логгере — это не создание экземпляра.
func (s *Svc) do() {
	s.logger.WithField("step", "start").Info("ok")
}
