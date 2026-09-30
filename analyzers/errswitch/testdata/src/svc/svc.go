// Eval of GID-279: a switch over an error variable sits inside `if err != nil`.
package svc

import (
	"context"
	"errors"
)

var (
	ErrRate     = errors.New("rate")
	ErrNoTable  = errors.New("no table")
	ErrRejected = errors.New("rejected")
)

type client struct{}

func (client) upload() (int, error) { return 0, nil }
func (client) failure() error       { return nil }

type result struct{ Err error }

type rateError struct{}

func (*rateError) Error() string { return "rate" }

// --- Positive: the incident — no `if err != nil` before the switch ---

func upload(c client) (int, error) {
	id, err := c.upload()

	switch { // want `GID-279: the error "err" is classified with a switch that is not inside .if err != nil.\. Fix: check the error with if first`
	case err == nil:
		return id, nil
	case errors.Is(err, ErrRate):
		return 0, ErrRejected
	default:
		return 0, err
	}
}

// --- Positive: errors.Is / errors.As without the guard ---

func classifyIs(c client) error {
	_, err := c.upload()

	switch { // want `GID-279: the error "err" is classified with a switch`
	case errors.Is(err, ErrRate):
		return ErrRejected
	case errors.Is(err, ErrNoTable):
		return nil
	}

	return err
}

func classifyAs(c client) bool {
	_, err := c.upload()
	var target *rateError

	switch { // want `GID-279: the error "err" is classified with a switch`
	case errors.As(err, &target):
		return true
	}

	return false
}

// --- Positive: a field, a compound condition, a tagged switch, a type switch ---

func field(res result) error {
	switch { // want `GID-279: the error "res.Err" is classified with a switch`
	case errors.Is(res.Err, ErrRate):
		return ErrRejected
	}

	return nil
}

func compound(c client, ready bool) error {
	_, err := c.upload()

	switch { // want `GID-279: the error "err" is classified with a switch`
	case ready && err == nil:
		return nil
	default:
		return err
	}
}

func tagged(c client) int {
	_, err := c.upload()

	switch err { // want `GID-279: the error "err" is classified with a switch`
	case ErrRate:
		return 1
	}

	return 0
}

func typed(c client) int {
	_, err := c.upload()

	switch e := err.(type) { // want `GID-279: the error "err" is classified with a switch`
	case *rateError:
		return len(e.Error())
	}

	return 0
}

// --- Positive: the switch is in a loop, or under an `if` that checks something else ---

func drain(c client) error {
	for {
		_, err := c.upload()

		switch { // want `GID-279: the error "err" is classified with a switch`
		case err == nil:
			continue
		case errors.Is(err, ErrNoTable):
			return nil
		default:
			return err
		}
	}
}

func otherGuard(c client, ready bool) error {
	_, err := c.upload()
	if ready {
		switch { // want `GID-279: the error "err" is classified with a switch`
		case errors.Is(err, ErrRate):
			return ErrRejected
		}
	}

	return err
}

// --- Positive: the else branch of `err != nil` is the success side, not a guard ---

func wrongBranch(c client) error {
	_, err := c.upload()
	if err != nil {
		return err
	} else {
		switch { // want `GID-279: the error "err" is classified with a switch`
		case errors.Is(err, ErrRate):
			return ErrRejected
		}
	}

	return nil
}

// --- Positive: a guard for another variable does not cover this one ---

func otherVariable(c client, prev error) error {
	_, err := c.upload()
	if prev != nil {
		switch { // want `GID-279: the error "err" is classified with a switch`
		case errors.Is(err, ErrRate):
			return ErrRejected
		}
	}

	return err
}

// --- Negative: the fix — the classes are narrowed inside `if err != nil` ---

func fixed(c client) (int, error) {
	id, err := c.upload()
	if err != nil {
		switch {
		case errors.Is(err, ErrRate):
			err = ErrRejected
		}

		return 0, err
	}

	return id, nil
}

// --- Negative: the guard in other forms ---

func initStatement(c client) error {
	if _, err := c.upload(); err != nil {
		switch {
		case errors.Is(err, ErrNoTable):
			return nil
		}

		return err
	}

	return nil
}

func conjunct(c client, ready bool) error {
	_, err := c.upload()
	if err != nil && ready {
		switch {
		case errors.Is(err, ErrRate):
			return ErrRejected
		}
	}

	return err
}

func elseOfNil(c client) error {
	_, err := c.upload()
	if err == nil {
		return nil
	} else {
		switch {
		case errors.Is(err, ErrRate):
			return ErrRejected
		}
	}

	return err
}

func parenthesised(c client) error {
	_, err := c.upload()
	if err != nil {
		switch {
		case errors.Is(err, ErrRate):
			return ErrRejected
		}
	}

	return err
}

// --- Negative: deeper inside the guard — a loop, a closure, a nested if ---

func nested(c client, items []int) error {
	_, err := c.upload()
	if err != nil {
		for range items {
			switch {
			case errors.Is(err, ErrRate):
				return ErrRejected
			}
		}

		handle := func() bool {
			switch e := err.(type) {
			case *rateError:
				return e != nil
			}

			return false
		}

		if handle() {
			return err
		}
	}

	return nil
}

func fieldGuard(res result) error {
	if res.Err != nil {
		switch {
		case errors.Is(res.Err, ErrRate):
			return ErrRejected
		}
	}

	return nil
}

// --- Negative: errors.Is / errors.As / ctx.Err() cases are fine in a switch, inside the guard ---

func classify(ctx context.Context, err error) int {
	var target *rateError
	if err != nil {
		switch {
		case errors.Is(err, ErrRate):
			return 1
		case errors.As(err, &target):
			return 2
		case ctx.Err() != nil:
			return 3
		}
	}

	return 0
}

// --- Boundary: a PARAMETER is not judged — the caller has checked the error ---

func errorResponse(err error) int {
	switch {
	case errors.Is(err, ErrRate):
		return 429
	case errors.Is(err, ErrNoTable):
		return 404
	}

	return 500
}

func isRetryable(err error) bool {
	switch err {
	case ErrRate:
		return true
	}

	return false
}

func (client) kind(err error) int {
	switch e := err.(type) {
	case *rateError:
		return len(e.Error())
	}

	return 0
}

func paramInClosure(c client) int {
	handle := func(err error) int {
		switch {
		case errors.Is(err, ErrRate):
			return 1
		}

		return 0
	}
	_, err := c.upload()
	if err != nil {
		return handle(err)
	}

	return 0
}

// --- Boundary: a call result compared with nil is not an error variable ---

func callResult(ctx context.Context, c client) int {
	switch {
	case ctx.Err() != nil:
		return 1
	case c.failure() != nil:
		return 2
	}

	return 0
}

// --- Boundary: a package-level sentinel alone, a non-error nil, an error only in a body ---

func sentinelOnly(kind error) int {
	switch {
	case ErrRate == ErrNoTable:
		return 1
	}

	return 0
}

func notError(p *int) int {
	switch {
	case p == nil:
		return 0
	default:
		return *p
	}
}

func errorInBody(kind int, err error) error {
	switch {
	case kind == 1:
		return err
	}

	return nil
}

func typeSwitchOnAny(v any) int {
	switch v.(type) {
	case int:
		return 1
	}

	return 0
}
