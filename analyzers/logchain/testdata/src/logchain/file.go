// Eval для GID-156 (цепочка logrus по вызову на строку).
package logchain

import (
	"context"

	"github.com/sirupsen/logrus"
)

type Svc struct {
	logger *logrus.Entry
}

// --- Позитивные кейсы ---

func (s *Svc) badInline(ctx context.Context, err error) {
	s.logger.WithContext(ctx).WithError(err).Error("failed") // want `GID-156: цепочка logrus оформляется по одному вызову на строке, включая первый`
}

// Граничный кейс: первый вызов прилип к ресиверу.
func (s *Svc) badFirstInline(ctx context.Context, err error) {
	s.logger.WithContext(ctx). // want `GID-156: цепочка logrus оформляется по одному вызову на строке, включая первый`
					WithError(err).
					Error("failed")
}

// Граничный кейс: два вызова на одной строке в середине цепочки.
func (s *Svc) badMiddle(ctx context.Context, err error) {
	s.logger.
		WithContext(ctx).WithError(err). // want `GID-156: цепочка logrus оформляется по одному вызову на строке, включая первый`
		Error("failed")
}

// --- Негативные кейсы ---

func (s *Svc) good(ctx context.Context, err error) {
	s.logger.
		WithContext(ctx).
		WithError(err).
		WithField("some", "field").
		Error("some text")
}

// Неприменимость: одиночный вызов — inline допустим.
func (s *Svc) single() {
	s.logger.Info("tick")
}
