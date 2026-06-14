// Eval for GID-156 (a logrus chain, one call per line).
package logchain

import (
	"context"

	"github.com/sirupsen/logrus"
)

type Svc struct {
	logger *logrus.Entry
}

// --- Positive cases ---

func (s *Svc) badInline(ctx context.Context, err error) {
	s.logger.WithContext(ctx).WithError(err).Error("failed") // want `GID-156: a logrus chain must put one call per line, including the first\. Fix: break each call onto a new line`
}

// Boundary case: the first call is stuck to the receiver.
func (s *Svc) badFirstInline(ctx context.Context, err error) {
	s.logger.WithContext(ctx). // want `GID-156: a logrus chain must put one call per line, including the first\. Fix: break each call onto a new line`
					WithError(err).
					Error("failed")
}

// Boundary case: two calls on one line in the middle of the chain.
func (s *Svc) badMiddle(ctx context.Context, err error) {
	s.logger.
		WithContext(ctx).WithError(err). // want `GID-156: a logrus chain must put one call per line, including the first\. Fix: break each call onto a new line`
		Error("failed")
}

// --- Negative cases ---

func (s *Svc) good(ctx context.Context, err error) {
	s.logger.
		WithContext(ctx).
		WithError(err).
		WithField("some", "field").
		Error("some text")
}

// Non-applicability: a single call — inline is allowed.
func (s *Svc) single() {
	s.logger.Info("tick")
}
