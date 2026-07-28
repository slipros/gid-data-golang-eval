// Eval for GID-253 (a logrus chain sets its fields in one WithFields call).
package logfields

import (
	"context"

	"github.com/sirupsen/logrus"
)

type Svc struct {
	logger *logrus.Entry
}

type Producer struct {
	logger logrus.FieldLogger
}

// --- Class 1: positive ---

// The reported case: the payload is set pair by pair.
func (s *Svc) badMany(ctx context.Context, err error, offset int64, level, retry int) {
	s.logger.
		WithContext(ctx).
		WithError(err).
		WithField("offset", offset). // want `GID-253: a logger chain sets its fields in 4 separate calls — they belong in one\. Fix: replace them with a single WithFields\(logrus\.Fields\{"offset": offset, "fallback_level": level\}\)`
		WithField("target_topic", s.topic(level)).
		WithField("fallback_level", level).
		WithField("fallback_retry", retry).
		Error("send fallback message")
}

// An inline chain — the same violation (the layout itself is GID-156).
func (s *Svc) badInline(a, b string) {
	s.logger.WithField("a", a).WithField("b", b).Info("done") // want `GID-253: a logger chain sets its fields in 2 separate calls`
}

// A logger interface behaves like the concrete entry.
func (p *Producer) badIface(err error, topic string, partition int32) {
	p.logger.
		WithError(err).
		WithField("topic", topic). // want `GID-253: a logger chain sets its fields in 2 separate calls`
		WithField("partition", partition).
		Error("publish")
}

// --- Class 2: negative ---

func (s *Svc) good(ctx context.Context, err error, offset int64, level int) {
	s.logger.
		WithContext(ctx).
		WithError(err).
		WithFields(logrus.Fields{
			"offset":         offset,
			"fallback_level": level,
		}).
		Error("send fallback message")
}

// A single pair is exactly what WithField is for.
func (s *Svc) goodSingle(ctx context.Context, offset int64) {
	s.logger.
		WithContext(ctx).
		WithField("offset", offset).
		Info("consumed")
}

// --- Class 3: boundary ---

// Exactly two field calls — the threshold.
func (s *Svc) badBoundary(a, b string) {
	s.logger.
		WithField("a", a). // want `GID-253: a logger chain sets its fields in 2 separate calls`
		WithField("b", b).
		Info("pair")
}

// A mix of the two methods is still two field calls.
func (s *Svc) badMixed(offset int64, level int) {
	s.logger.
		WithFields(logrus.Fields{"offset": offset}). // want `GID-253: a logger chain sets its fields in 2 separate calls`
		WithField("fallback_level", level).
		Info("mixed")
}

// A chain without a terminal call: the entry is stored in a variable.
func (s *Svc) badNoTerminal(a, b string) *logrus.Entry {
	return s.logger.
		WithField("a", a). // want `GID-253: a logger chain sets its fields in 2 separate calls`
		WithField("b", b)
}

// Fields attached through separate statements are outside a chain — the rule
// does not chase an entry across statements.
func (s *Svc) goodSeparateStatements(a, b string) {
	entry := s.logger.WithField("a", a)
	entry = entry.WithField("b", b)
	entry.Info("stepwise")
}

// --- Class 4: non-applicability ---

type fakeLogger struct{}

func (f *fakeLogger) WithField(key string, value any) *fakeLogger { return f }
func (f *fakeLogger) Info(args ...any)                            {}

// Not a logrus type — the rule says nothing about it.
func notALogger(f *fakeLogger, a, b string) {
	f.WithField("a", a).WithField("b", b).Info("custom")
}

// WithContext/WithError are not field calls.
func (s *Svc) goodContextAndError(ctx context.Context, err error) {
	s.logger.
		WithContext(ctx).
		WithError(err).
		Error("failed")
}

func (s *Svc) topic(level int) string { return "topic" }
