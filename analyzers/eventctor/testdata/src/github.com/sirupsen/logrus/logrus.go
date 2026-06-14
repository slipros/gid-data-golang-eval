// Stub of github.com/sirupsen/logrus for eval.
package logrus

type Fields map[string]any

type Entry struct{}

func (e *Entry) WithField(key string, value any) *Entry { return e }
func (e *Entry) Info(args ...any)                        {}

type Logger struct{}

func (l *Logger) WithField(key string, value any) *Entry { return &Entry{} }
func (l *Logger) Info(args ...any)                        {}
