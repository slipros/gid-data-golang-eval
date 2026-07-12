// Stub of github.com/pkg/errors for eval. Mirrors the real package's
// Is/As forwarders to the standard library (present since v0.9.1) — the
// point being that their declaring package path is github.com/pkg/errors,
// not "errors".
package errors

import stderrors "errors"

func New(message string) error { return stderrors.New(message) }

func Errorf(format string, args ...any) error { return stderrors.New(format) }

func Wrap(err error, message string) error { return err }

func Wrapf(err error, format string, args ...any) error { return err }

func WithStack(err error) error { return err }

func WithMessage(err error, message string) error { return err }

func WithMessagef(err error, format string, args ...any) error { return err }

// Is forwards to the standard library errors.Is (pkg/errors v0.9.1+).
func Is(err, target error) bool { return stderrors.Is(err, target) }

// As forwards to the standard library errors.As (pkg/errors v0.9.1+).
func As(err error, target any) bool { return stderrors.As(err, target) }
