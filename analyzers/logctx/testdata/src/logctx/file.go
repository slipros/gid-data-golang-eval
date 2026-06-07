// Eval для GID-155 (WithContext / WithError).
package logctx

import (
	"context"

	"github.com/sirupsen/logrus"
)

type Svc struct {
	logger *logrus.Entry
}

// --- Позитивные кейсы ---

func (s *Svc) badNoCtx(ctx context.Context) {
	s.logger.Info("start") // want `GID-155: лог-вызов в функции с ctx обязан содержать WithContext\(ctx\)`
}

func (s *Svc) badErrorNoErr(ctx context.Context, err error) {
	s.logger.
		WithContext(ctx).
		Error("failed") // want `GID-155: лог уровня Error обязан содержать WithError\(err\)`
}

// Граничный кейс: оба нарушения в одном вызове.
func (s *Svc) badBoth(ctx context.Context) {
	s.logger.Error("failed") // want `GID-155: лог-вызов в функции с ctx обязан содержать WithContext\(ctx\)` `GID-155: лог уровня Error обязан содержать WithError\(err\)`
}

// Граничный кейс: ctx есть у внешней функции, но лог внутри замыкания
// без ctx — требование к замыканию не предъявляется.
func (s *Svc) closure(ctx context.Context) func() {
	return func() {
		s.logger.Info("tick")
	}
}

// --- Негативные кейсы ---

func (s *Svc) good(ctx context.Context, err error) {
	s.logger.
		WithContext(ctx).
		WithError(err).
		Error("failed")
}

func (s *Svc) goodInfo(ctx context.Context) {
	s.logger.
		WithContext(ctx).
		WithField("step", "start").
		Info("start")
}

// --- Неприменимость: функция без ctx, уровень не Error ---

func (s *Svc) notApplicable() {
	s.logger.Info("no ctx required")
}
