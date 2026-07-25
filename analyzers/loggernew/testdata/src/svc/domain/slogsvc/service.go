// Eval for GID-214 on the slog stack: slog.New/Default/SetDefault outside the
// composition root are the same smell as logrus.New/StandardLogger.
package slogsvc

import (
	"log/slog"
	"os"
)

type Service struct {
	logger *slog.Logger
}

// --- Positive: the logger is created/grabbed outside the composition root ---

func NewService() *Service {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil)) // want `GID-214: slog\.New\(\) may be called only in the composition root \(main, internal/app\)\. Fix: pass a ready logger \(\*logrus\.Entry, \*slog\.Logger\) through the constructor`

	return &Service{logger: logger}
}

func NewFromDefault() *Service {
	return &Service{logger: slog.Default()} // want `GID-214: slog\.Default\(\) may be called only in the composition root \(main, internal/app\)\. Fix: pass a ready logger \(\*logrus\.Entry, \*slog\.Logger\) through the constructor`
}

func Init() {
	slog.SetDefault(slog.Default()) // want `GID-214: slog\.SetDefault\(\) may be called only in the composition root \(main, internal/app\)\. Fix: pass a ready logger \(\*logrus\.Entry, \*slog\.Logger\) through the constructor` `GID-214: slog\.Default\(\) may be called only in the composition root \(main, internal/app\)\. Fix: pass a ready logger \(\*logrus\.Entry, \*slog\.Logger\) through the constructor`
}

// --- Negative: a ready logger arrives through the constructor ---

func NewServiceWithLogger(logger *slog.Logger) *Service {
	return &Service{logger: logger.With("service", "slogsvc")}
}
