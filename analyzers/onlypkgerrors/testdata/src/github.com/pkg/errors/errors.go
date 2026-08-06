// A stub of github.com/pkg/errors for the eval.
package errors

import stderrors "errors"

func New(message string) error { return stderrors.New(message) }

func Errorf(format string, args ...any) error { return stderrors.New(format) }

func Wrap(err error, message string) error { return err }

func Wrapf(err error, format string, args ...any) error { return err }

func WithStack(err error) error { return err }

func WithMessage(err error, message string) error { return err }

// Is/As/Unwrap — the re-exports of go113.go (v0.9.0+): one-line delegates to the
// std functions. Their presence is what makes the std errors import redundant.

func Is(err, target error) bool { return stderrors.Is(err, target) }

func As(err error, target any) bool { return stderrors.As(err, target) }

func Unwrap(err error) error { return stderrors.Unwrap(err) }
