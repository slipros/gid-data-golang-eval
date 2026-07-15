// Eval of GID-244: map a boundary error to a sentinel by reassign-then-wrap-once,
// not by wrapping the sentinel in a guard branch that duplicates the outer Wrap.
package repository

import (
	"github.com/pkg/errors"

	"svc/dal/entity"
)

type Conn interface {
	call() error
}

type Repo struct {
	conn Conn
}

func isNoResult(err error) bool  { return err != nil }
func isRetryable(err error) bool { return err != nil }

// --- Class 1: positive ---

func (r *Repo) badSentinelGuard() error {
	err := r.conn.call()
	if err != nil {
		if isNoResult(err) { // want `GID-244: a sentinel wrapped in a guard branch duplicates the outer errors\.Wrap`
			return errors.Wrap(entity.ErrNoResult, "update key")
		}
		return errors.Wrap(err, "update key")
	}
	return nil
}

func (r *Repo) badSentinelIs() error {
	err := r.conn.call()
	if err != nil {
		if errors.Is(err, entity.ErrNoResult) { // want `GID-244: a sentinel wrapped in a guard branch duplicates the outer errors\.Wrap`
			return errors.Wrap(entity.ErrNoResult, "op")
		}
		return errors.Wrap(err, "op")
	}
	return nil
}

// --- Class 2: negative ---

// The canonical shape GID-244 pushes toward: reassign, then one shared Wrap.
func (r *Repo) goodReassign() error {
	err := r.conn.call()
	if err != nil {
		if isNoResult(err) {
			err = entity.ErrNoResult
		}
		return errors.Wrap(err, "update key")
	}
	return nil
}

// Distinct context messages — not a pure dedup, left alone.
func (r *Repo) goodDiffMsg() error {
	err := r.conn.call()
	if err != nil {
		if isNoResult(err) {
			return errors.Wrap(entity.ErrNoResult, "no rows")
		}
		return errors.Wrap(err, "exec failed")
	}
	return nil
}

// Both branches wrap err — there is no sentinel to reassign.
func (r *Repo) goodGuardWrapsErr() error {
	err := r.conn.call()
	if err != nil {
		if isRetryable(err) {
			return errors.Wrap(err, "op")
		}
		return errors.Wrap(err, "op")
	}
	return nil
}

// --- Class 3: boundary ---

// The guard does more than map-and-return — a mechanical collapse would drop the extra statement.
func (r *Repo) boundaryExtraStmt() error {
	err := r.conn.call()
	if err != nil {
		if isNoResult(err) {
			_ = err
			return errors.Wrap(entity.ErrNoResult, "op")
		}
		return errors.Wrap(err, "op")
	}
	return nil
}

// The predicate tests an unrelated flag, not err — cannot become a reassignment of err.
func (r *Repo) boundaryFlag(useSentinel bool) error {
	err := r.conn.call()
	if err != nil {
		if useSentinel {
			return errors.Wrap(entity.ErrNoResult, "op")
		}
		return errors.Wrap(err, "op")
	}
	return nil
}

// --- Class 4: non-applicability ---

// No mirror errors.Wrap(err, ...) beside the guard.
func (r *Repo) naNoMirror() error {
	err := r.conn.call()
	if err != nil {
		if isNoResult(err) {
			return errors.Wrap(entity.ErrNoResult, "op")
		}
		return errors.WithStack(err)
	}
	return nil
}

// Wrapf (formatted) is out of scope in v1.
func (r *Repo) naWrapf(n int) error {
	err := r.conn.call()
	if err != nil {
		if isNoResult(err) {
			return errors.Wrapf(entity.ErrNoResult, "op %d", n)
		}
		return errors.Wrapf(err, "op %d", n)
	}
	return nil
}
