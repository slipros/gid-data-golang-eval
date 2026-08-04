// Eval of GID-176 (part 1): the /dal/repository boundary.
// The boundary is an interface-method call (an injected external dependency).
package repository

import (
	"github.com/pkg/errors"

	"boundarysvc/dal/entity"
)

// Conn is the injected external dependency — calls to its methods are the boundary.
type Conn interface {
	call() error
	callRow() (int, error)
}

// buildQuery is a LOCAL pure helper (a package function), not the boundary.
func buildQuery() (string, error) { return "", nil }

type Repo struct {
	conn Conn
}

// concreteHelper is a method on a concrete type, not an interface — not the boundary.
func (r *Repo) concreteHelper() error { return nil }

// --- Positive: pass-through of a non-static error from an interface-method call ---

func (r *Repo) badPassThrough() error {
	err := r.conn.call()
	return err // want `GID-176: an error from an external call must be wrapped with errors\.Wrap\. Fix: collect stack and context`
}

func (r *Repo) badPassThroughMulti() (int, error) {
	n, err := r.conn.callRow()
	return n, err // want `GID-176: an error from an external call must be wrapped with errors\.Wrap\. Fix: collect stack and context`
}

// --- Positive: WithStack/WithMessage add no context ---

func (r *Repo) badWithStack() error {
	err := r.conn.call()
	return errors.WithStack(err) // want `GID-176: an error from an external call must be wrapped with errors\.Wrap \(WithStack adds no context\)`
}

func (r *Repo) badWithMessage() error {
	err := r.conn.call()
	return errors.WithMessage(err, "ctx") // want `GID-176: an error from an external call must be wrapped with errors\.Wrap \(WithMessage adds no context\)`
}

// --- Positive: a TRANSIT helper around the boundary error (the 2026-08-04
// incident). A mapper/normalizer takes the error and hands it back without a
// stack, so the return is still a pass-through — wrapping it in WithMessage
// used to hide it from this rule entirely, because the wrapper's argument
// stopped being a plain identifier. ---

// mapError is the transit helper: one error in, one error out (GID-242 bans
// the shape itself; here it only has to stay visible to GID-176).
func mapError(err error) error { return err }

func (r *Repo) badTransitWithMessage() error {
	err := r.conn.call()
	return errors.WithMessage(mapError(err), "select") // want `GID-176: an error from an external call must be wrapped with errors\.Wrap \(WithMessage adds no context\)`
}

func (r *Repo) badTransitWithStack() error {
	err := r.conn.call()
	return errors.WithStack(mapError(err)) // want `GID-176: an error from an external call must be wrapped with errors\.Wrap \(WithStack adds no context\)`
}

func (r *Repo) badTransitNested() error {
	err := r.conn.call()
	return errors.WithMessage(mapError(mapError(err)), "select") // want `GID-176: an error from an external call must be wrapped with errors\.Wrap \(WithMessage adds no context\)`
}

func (r *Repo) badTransitBare() error {
	err := r.conn.call()
	return mapError(err) // want `GID-176: an error from an external call must be wrapped with errors\.Wrap \(passing it through a helper adds no stack\)`
}

// --- Negative: the transit chain already collects a stack — unwrapping stops
// at Wrap/WithStack, so enriching the result with a message is legitimate. ---

func (r *Repo) goodMessageOverWrap() error {
	err := r.conn.call()
	return errors.WithMessage(errors.Wrap(err, "select"), "ctx")
}

// --- Negative: a call that CONSUMES the error rather than forwarding it (more
// than one argument) is not a transit — nothing to look through. ---

func combine(err error, msg string) error { return err }

func (r *Repo) goodConsumingCall() error {
	err := r.conn.call()
	return errors.Wrap(combine(err, "x"), "select")
}

// --- Negative: the error from an interface-method call is wrapped with Wrap ---

func (r *Repo) goodWrap() error {
	err := r.conn.call()
	return errors.Wrap(err, "select")
}

// --- Negative: a local package function (a pure builder) is not the boundary ---
// Its error may be enriched with WithMessage (no second stack) or passed through.

func (r *Repo) goodLocalWithMessage() error {
	_, err := buildQuery()
	if err != nil {
		return errors.WithMessage(err, "build query")
	}
	return nil
}

func (r *Repo) goodLocalPassThrough() error {
	_, err := buildQuery()
	return err
}

// --- Negative: map a boundary error to a sentinel, then a single Wrap ---
// Reassigning err to a sentinel before one errors.Wrap avoids wrapping twice.

func isNoResult(err error) bool { return err != nil }

func (r *Repo) goodSentinelThenWrap() error {
	err := r.conn.call()
	if isNoResult(err) {
		err = entity.ErrNotFound
	}
	return errors.Wrap(err, "select")
}

// --- Negative: a method on a concrete type is not an interface-method call ---

func (r *Repo) goodConcreteWithMessage() error {
	err := r.concreteHelper()
	return errors.WithMessage(err, "helper")
}

// --- Boundary: returning a static error (GID-177 territory, not GID-176) ---

func (r *Repo) goodStatic() error {
	err := r.conn.call()
	if err != nil {
		return entity.ErrNotFound
	}
	return nil
}

// --- Inapplicable: the function does not return an error ---

func (r *Repo) noError() int {
	_ = r.conn.call()
	return 0
}
