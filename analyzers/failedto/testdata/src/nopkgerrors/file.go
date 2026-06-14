package nopkgerrors

import stderrors "errors"

// --- Class 4: inapplicability — a file without github.com/pkg/errors ---
// A local Wrap function with the same name must not be matched.

func Wrap(err error, message string) error { return err }

func useLocalWrap(err error) error {
	return Wrap(err, "failed to select") // not pkg/errors — not matched
}

func useStd() error {
	return stderrors.New("failed to do") // std — GID-146 territory, not matched
}
