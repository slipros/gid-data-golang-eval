// Stub github.com/pkg/errors для eval GID-186.
package errors

import stderrors "errors"

func New(message string) error { return stderrors.New(message) }

func Errorf(format string, args ...any) error { return stderrors.New(format) }

func Wrap(err error, message string) error { return err }

func Wrapf(err error, format string, args ...any) error { return err }

func WithMessage(err error, message string) error { return err }

func WithMessagef(err error, format string, args ...any) error { return err }
