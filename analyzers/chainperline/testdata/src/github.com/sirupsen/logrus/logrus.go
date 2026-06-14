// Logrus stub for the GID-196 eval: logrus chains are the domain of GID-156.
package logrus

type Entry struct{}

func (e *Entry) WithField(key string, value any) *Entry { return e }

func (e *Entry) Info(args ...any) {}

type Logger struct{}

func (l *Logger) WithField(key string, value any) *Entry { return &Entry{} }
