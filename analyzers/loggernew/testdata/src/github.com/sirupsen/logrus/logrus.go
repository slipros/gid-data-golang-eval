// Stub github.com/sirupsen/logrus для eval GID-214.
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

// New создаёт новый экземпляр логгера — package-level функция.
func New() *Logger { return &Logger{} }

// StandardLogger возвращает глобальный экземпляр логгера.
func StandardLogger() *Logger { return &Logger{} }
