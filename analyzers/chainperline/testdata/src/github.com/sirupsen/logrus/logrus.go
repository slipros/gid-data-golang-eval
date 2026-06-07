// Стаб logrus для eval GID-196: цепочки logrus — зона GID-156.
package logrus

type Entry struct{}

func (e *Entry) WithField(key string, value any) *Entry { return e }

func (e *Entry) Info(args ...any) {}

type Logger struct{}

func (l *Logger) WithField(key string, value any) *Entry { return &Entry{} }
