// Eval GID-214: positive — logrus.New() in /domain/service.
package service

import (
	"github.com/sirupsen/logrus"
)

type Svc struct {
	logger *logrus.Entry
}

// Positive case: creating a logger in the service is forbidden.
func New() *Svc {
	l := logrus.New() // want `GID-214: logrus.New\(\) may be called only in the composition root \(main, internal/app\)\. Fix: pass a ready \*logrus\.Entry through the constructor`
	_ = l
	return &Svc{}
}

// Negative: WithField on a ready logger is not instance creation.
func (s *Svc) do() {
	s.logger.WithField("step", "start").Info("ok")
}
