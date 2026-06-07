// Stub github.com/sirupsen/logrus для eval GID-181.
package logrus

type Fields map[string]any

type Entry struct{}

func (e *Entry) WithError(err error) *Entry             { return e }
func (e *Entry) WithField(key string, value any) *Entry { return e }
func (e *Entry) Info(args ...any)                       {}
func (e *Entry) Error(args ...any)                      {}
func (e *Entry) Fatal(args ...any)                      {}
func (e *Entry) Fatalf(format string, args ...any)      {}
func (e *Entry) Fatalln(args ...any)                    {}

type Logger struct{}

func (l *Logger) WithField(key string, value any) *Entry { return &Entry{} }
func (l *Logger) Info(args ...any)                       {}
func (l *Logger) Fatal(args ...any)                      {}
func (l *Logger) Fatalf(format string, args ...any)      {}
func (l *Logger) Fatalln(args ...any)                    {}

// Пакетные функции.
func Info(args ...any)                  {}
func Fatal(args ...any)                 {}
func Fatalf(format string, args ...any) {}
func Fatalln(args ...any)               {}
func Exit(code int)                     {}
