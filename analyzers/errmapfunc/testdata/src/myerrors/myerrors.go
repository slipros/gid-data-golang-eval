// Stub of a project-internal errors facade (re-exports Is/As) — stands in
// for a package a project would add via settings.packages. Its import path is
// neither "errors" nor "github.com/pkg/errors", so the DEFAULT whitelist does
// not cover it; only a custom settings.packages does.
package myerrors

import stderrors "errors"

func New(message string) error { return stderrors.New(message) }

func Is(err, target error) bool { return stderrors.Is(err, target) }

func As(err error, target any) bool { return stderrors.As(err, target) }
