// Eval of GID-244, the switch shape: several case clauses wrapping their own error
// with the same message.
package repository

import (
	"github.com/pkg/errors"

	"svc/dal/entity"
)

// --- Positive: the incident — four clauses, three of them share one message ---

func (r *Repo) badSwitch() (int, error) {
	err := r.conn.call()
	switch {
	case isNoResult(err): // want `GID-244: 3 case clauses wrap their own error with the same message "upload"\. Fix: assign err in each case`
		return 0, errors.Wrap(&entity.CabinetError{Reason: "a"}, "upload")
	case isRetryable(err):
		return 0, errors.Wrap(&entity.CabinetError{Reason: "b"}, "upload")
	case err != nil && err.Error() == "x":
		return 0, errors.Wrap(entity.ErrNoResult, "upload")
	case err != nil && err.Error() == "y":
		return 0, errors.Wrap(entity.ErrNoResult, "other")
	default:
		return 0, err
	}
}

// --- Positive: a tagged switch, a single result ---

func (r *Repo) badTagged(kind int) error {
	switch kind {
	case 1: // want `GID-244: 2 case clauses wrap their own error with the same message "op"\. Fix: assign err in each case`
		return errors.Wrap(entity.ErrNoResult, "op")
	case 2:
		return errors.Wrap(&entity.CabinetError{}, "op")
	}

	return nil
}

// --- Negative: the fix — assign per case, wrap once ---

func (r *Repo) goodSwitch() (int, error) {
	err := r.conn.call()
	switch {
	case isNoResult(err):
		err = &entity.CabinetError{Reason: "a"}
	case isRetryable(err):
		err = entity.ErrNoResult
	}
	if err != nil {
		return 0, errors.Wrap(err, "upload")
	}

	return 1, nil
}

// --- Boundary: only one clause per message — nothing duplicated ---

func (r *Repo) distinctMessages(kind int) error {
	switch kind {
	case 1:
		return errors.Wrap(entity.ErrNoResult, "first")
	case 2:
		return errors.Wrap(entity.ErrNoResult, "second")
	}

	return nil
}

// --- Boundary: clauses wrapping err itself are not sentinel mapping ---

func (r *Repo) wrapsErr(kind int) error {
	err := r.conn.call()
	switch kind {
	case 1:
		return errors.Wrap(err, "op")
	case 2:
		return errors.Wrap(err, "op")
	}

	return nil
}

// --- Boundary: a clause with more than the one return is not judged ---

func (r *Repo) extraStatement(kind int) error {
	switch kind {
	case 1:
		kind++
		return errors.Wrap(entity.ErrNoResult, "op")
	case 2:
		kind--
		return errors.Wrap(entity.ErrNoResult, "op")
	}

	return nil
}
