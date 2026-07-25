// Eval of the built-in GID-217 ban on the logrus logger interfaces: they have
// no WithContext, so a dependency typed by them cannot satisfy GID-155.
package logrusiface

import "github.com/sirupsen/logrus"

// --- Positive: a dependency typed by a logrus interface ---

type Service struct {
	logger logrus.FieldLogger // want `GID-217: logrus\.FieldLogger has no WithContext — a log call cannot carry the context \(GID-155\)\. Fix: type the dependency as \*logrus\.Entry \(or \*slog\.Logger\)`
}

func NewStd(logger logrus.StdLogger) {} // want `GID-217: logrus\.StdLogger has no WithContext and no fields — a log call cannot carry the context \(GID-155\)\. Fix: type the dependency as \*logrus\.Entry \(or \*slog\.Logger\)`

func NewExt(logger logrus.Ext1FieldLogger) {} // want `GID-217: logrus\.Ext1FieldLogger has no WithContext — a log call cannot carry the context \(GID-155\)\. Fix: type the dependency as \*logrus\.Entry \(or \*slog\.Logger\)`

// --- Negative: the concrete *logrus.Entry carries the context ---

type Good struct {
	logger *logrus.Entry
}

func NewGood(logger *logrus.Entry) *Good {
	return &Good{logger: logger.WithField("good", "svc")}
}
