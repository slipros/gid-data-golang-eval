// Stub of github.com/sirupsen/logrus for eval.
package logrus

import "context"

type Entry struct{}

func (e *Entry) WithContext(ctx context.Context) *Entry { return e }

func (e *Entry) WithField(key string, value any) *Entry { return e }

func (e *Entry) Info(args ...any) {}

// FieldLogger — the interface without WithContext.
type FieldLogger interface {
	WithField(key string, value any) *Entry
	Info(args ...any)
}

// StdLogger — the bare stdlib-shaped interface.
type StdLogger interface {
	Print(args ...any)
}

// Ext1FieldLogger — FieldLogger plus the *ln methods.
type Ext1FieldLogger interface {
	FieldLogger
	Infoln(args ...any)
}
