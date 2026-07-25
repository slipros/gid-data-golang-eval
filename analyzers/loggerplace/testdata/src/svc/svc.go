// Eval of GID-251: the logger is a dependency, not configuration — it must not
// live in an options struct.
package svc

import (
	"log/slog"
	"time"

	"github.com/sirupsen/logrus"
)

// --- Positive: a logger inside an options struct ---

type SinkOptions struct {
	// Dir — where the reports are written.
	Dir string

	// Logger — the application logger.
	Logger *slog.Logger // want `GID-251: options struct "SinkOptions" holds a logger — a logger is a dependency, not configuration\. Fix: drop the field and take the logger as a separate constructor parameter, after opts`
}

type CacheConfig struct {
	TTL    time.Duration
	Logger *logrus.Entry // want `GID-251: options struct "CacheConfig" holds a logger — a logger is a dependency, not configuration\. Fix: drop the field and take the logger as a separate constructor parameter, after opts`
}

// Boundary: an interface-typed logger field is a logger too.
type ReporterSettings struct {
	Logger logrus.FieldLogger // want `GID-251: options struct "ReporterSettings" holds a logger — a logger is a dependency, not configuration\. Fix: drop the field and take the logger as a separate constructor parameter, after opts`
}

// --- Negative: options without a logger, the logger comes as a parameter ---

type SenderOptions struct {
	Dir     string
	Retries int
}

type Sender struct {
	opts   SenderOptions
	logger *slog.Logger
}

func NewSender(opts SenderOptions, logger *slog.Logger) *Sender {
	return &Sender{opts: opts, logger: logger.With(slog.String("sender", "mail"))}
}

// --- Boundary: a struct without the options suffix keeps its logger field ---

type Sink struct {
	logger *slog.Logger
}

// --- Non-applicability: an options struct of plain configuration ---

type PlainOptions struct {
	Timeout time.Duration
	Debug   bool
}
