package svc

import (
	stderrors "errors"
	"fmt"

	"github.com/pkg/errors"
)

// --- Class 1: positive (the violation is caught) ---

// A var with New + the "Failed" prefix (case-insensitive).
var ErrSelect = errors.New("Failed: x") // want `GID-184: error message starts with "failed"`

func wrapFailed(err error) error {
	return errors.Wrap(err, "failed to select") // want `GID-184: error message starts with "failed to"`
}

func withMessageUnable(err error) error {
	return errors.WithMessage(err, "unable to parse") // want `GID-184: error message starts with "unable to"`
}

func errorfError(id int) error {
	return errors.Errorf("error while loading %d", id) // want `GID-184: error message starts with "error"`
}

func wrapfCannot(err error, id int) error {
	return errors.Wrapf(err, "cannot save %d", id) // want `GID-184: error message starts with "cannot"`
}

func withMessagefCouldNot(err error, id int) error {
	return errors.WithMessagef(err, "could not commit %d", id) // want `GID-184: error message starts with "could not"`
}

// --- Class 2: negative (clean code passes) ---

func wrapClean(err error) error {
	return errors.Wrap(err, "select user")
}

func newClean() error {
	return errors.New("parse config")
}

// --- Class 3: boundary (looks like a violation, but is acceptable) ---

// "failure mode" — the word failure is not in the list, the word boundary protects it.
func wrapFailureMode(err error) error {
	return errors.Wrap(err, "failure mode handling")
}

// fmt.Sprintf — a different package, not matched.
func sprintfNotMatched(err error) error {
	return errors.Wrap(err, fmt.Sprintf("%s", "x"))
}

// std errors.New — that is GID-146 territory, not GID-184.
func stdErrorsNew() error {
	return stderrors.New("failed to do thing")
}

// A non-literal message (a variable) — not matched.
func wrapVariable(err error, msg string) error {
	return errors.Wrap(err, msg)
}

// Concatenation with a variable — not a literal, not matched.
func wrapConcat(err error, name string) error {
	return errors.Wrap(err, "failed to "+name)
}
