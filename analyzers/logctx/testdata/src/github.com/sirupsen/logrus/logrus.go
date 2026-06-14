// Stub github.com/sirupsen/logrus for eval.
package logrus

import "context"

type Fields map[string]any

type Entry struct{}

func (e *Entry) WithContext(ctx context.Context) *Entry      { return e }
func (e *Entry) WithError(err error) *Entry                  { return e }
func (e *Entry) WithField(key string, value any) *Entry      { return e }
func (e *Entry) WithFields(fields Fields) *Entry             { return e }
func (e *Entry) Trace(args ...any)                           {}
func (e *Entry) Debug(args ...any)                           {}
func (e *Entry) Info(args ...any)                            {}
func (e *Entry) Infof(format string, args ...any)            {}
func (e *Entry) Warn(args ...any)                            {}
func (e *Entry) Error(args ...any)                           {}
func (e *Entry) Errorf(format string, args ...any)           {}

type Logger struct{}

func (l *Logger) WithContext(ctx context.Context) *Entry { return &Entry{} }
func (l *Logger) WithError(err error) *Entry              { return &Entry{} }
func (l *Logger) WithField(key string, value any) *Entry  { return &Entry{} }
func (l *Logger) Info(args ...any)                        {}
func (l *Logger) Error(args ...any)                       {}

type FieldLogger interface {
	WithError(err error) *Entry
	WithField(key string, value any) *Entry
	Info(args ...any)
	Error(args ...any)
}
