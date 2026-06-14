// Stub github.com/sirupsen/logrus for the GID-214 eval.
package logrus

import "context"

type Fields map[string]any

type Entry struct{}

func (e *Entry) WithContext(ctx context.Context) *Entry { return e }
func (e *Entry) WithError(err error) *Entry             { return e }
func (e *Entry) WithField(key string, value any) *Entry { return e }
func (e *Entry) Info(args ...any)                       {}
func (e *Entry) Error(args ...any)                      {}

type Logger struct{}

func (l *Logger) WithField(key string, value any) *Entry { return &Entry{} }
func (l *Logger) Info(args ...any)                       {}

// New creates a new logger instance — a package-level function.
func New() *Logger { return &Logger{} }

// StandardLogger returns the global logger instance.
func StandardLogger() *Logger { return &Logger{} }
