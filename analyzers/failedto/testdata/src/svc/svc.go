package svc

import (
	stderrors "errors"
	"fmt"

	"github.com/pkg/errors"
)

// --- Класс 1: позитивный (нарушение ловится) ---

// var с New + префикс "Failed" (регистронезависимо).
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

// --- Класс 2: негативный (чистый код проходит) ---

func wrapClean(err error) error {
	return errors.Wrap(err, "select user")
}

func newClean() error {
	return errors.New("parse config")
}

// --- Класс 3: граничный (похоже на нарушение, но допустимо) ---

// "failure mode" — слово failure не входит в список, граница слова защищает.
func wrapFailureMode(err error) error {
	return errors.Wrap(err, "failure mode handling")
}

// fmt.Sprintf — другой пакет, не матчится.
func sprintfNotMatched(err error) error {
	return errors.Wrap(err, fmt.Sprintf("%s", "x"))
}

// std errors.New — это зона GID-146, не GID-184.
func stdErrorsNew() error {
	return stderrors.New("failed to do thing")
}

// Не-литеральное сообщение (переменная) — не матчим.
func wrapVariable(err error, msg string) error {
	return errors.Wrap(err, msg)
}

// Конкатенация с переменной — не литерал, не матчим.
func wrapConcat(err error, name string) error {
	return errors.Wrap(err, "failed to "+name)
}
