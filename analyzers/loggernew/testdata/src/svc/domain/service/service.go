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
	l := logrus.New() // want `GID-214: logrus.New\(\) вызывается только в composition root \(main, internal/app\) — пробрасывай готовый \*logrus\.Entry через конструктор`
	_ = l
	return &Svc{}
}

// Негатив: WithField на готовом логгере — это не создание экземпляра.
func (s *Svc) do() {
	s.logger.WithField("step", "start").Info("ok")
}
