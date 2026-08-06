// Eval of GID-256: the TEXT of an error (err.Error()) is not passed into an
// error constructor — flattening keeps the text and drops the chain and the
// stack. The fixtures mirror the 2026-08-06 incident (ad-cabinet-connector):
// the client half smuggles the cause into a sentinel's message, the service
// half carries it across a sentinel replacement in a local variable.
package svc

import (
	"fmt"
	"log"

	"gderror"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrServerError is the package-level sentinel the incident flattened into.
var ErrServerError = errors.New("server error")

// ErrRetryable stands in for the domain sentinel the consumer replaced with.
var ErrRetryable = errors.New("retryable")

// ServerError is the typed error the rule points at: a stable class that can
// carry the cause.
type ServerError struct {
	Err error
}

func (e *ServerError) Error() string { return "server error: " + e.Err.Error() }

func (e *ServerError) Unwrap() error { return e.Err }

// Conn is the injected dependency standing in for an HTTP client.
type Conn interface {
	Do() error
}

// Client is the boundary client of the fixtures below.
type Client struct {
	conn Conn
}

// --- Positive: the incident shape — the wrapped cause is used only for its
// text, and the stack Wrapf collected on the line above goes with it. ---

func (c *Client) Confirm(id int64) error {
	if err := c.conn.Do(); err != nil {
		wrapped := errors.Wrapf(err, "confirm segment %d", id)

		return errors.WithMessage(ErrServerError, wrapped.Error()) // want `GID-256: the text of an error is passed into an error constructor`
	}
	return nil
}

// --- Positive: the consumer half — the text is carried across the sentinel
// replacement in a local, so the constructor call itself looks innocent. ---

func (c *Client) Deliver() error {
	err := c.conn.Do()
	if err != nil {
		msg := err.Error()
		err = ErrRetryable

		return errors.Wrap(err, msg) // want `GID-256: the text of an error is passed into an error constructor`
	}
	return nil
}

// --- Positive: errors.New over a cause's text — the flattening at its most literal. ---

func newFromText(err error) error {
	return errors.New(err.Error()) // want `GID-256: the text of an error is passed into an error constructor`
}

// --- Positive: fmt.Errorf with %s instead of %w. ---

func formatText(err error) error {
	return fmt.Errorf("do x: %s", err.Error()) // want `GID-256: the text of an error is passed into an error constructor`
}

// --- Positive: the text hides among the format arguments of WithMessagef. ---

func messagefText(err error, code int) error {
	return errors.WithMessagef(ErrServerError, "status %d: %s", code, err.Error()) // want `GID-256: the text of an error is passed into an error constructor`
}

// --- Negative: the cause goes into the chain. ---

func (c *Client) ConfirmClean(id int64) error {
	if err := c.conn.Do(); err != nil {
		return errors.Wrapf(err, "confirm segment %d", id)
	}
	return nil
}

// --- Negative: a typed error carries the cause, Wrap collects the stack once —
// the shape the diagnostic asks for. ---

func (c *Client) ConfirmTyped(id int64) error {
	if err := c.conn.Do(); err != nil {
		return errors.Wrapf(&ServerError{Err: err}, "confirm segment %d", id)
	}
	return nil
}

// --- Negative: err.Error() outside an error constructor — a gRPC status
// message is a string by contract, nothing is flattened into an error. ---

func toStatus(err error) *status.Status {
	return status.New(codes.Internal, err.Error())
}

// --- Negative: a log field. ---

func logError(err error) {
	log.Println("failed:", err.Error())
}

// --- Boundary: the first argument of Wrap is the error being wrapped, not a message. ---

func wrapPlain(err error) error {
	return errors.Wrap(err, "ctx")
}

// --- Boundary: a variable holding the text but never handed to a constructor. ---

func textOnly(err error) string {
	msg := err.Error()

	return msg
}

// --- Boundary: reassigned from something else — the value at the constructor
// is unknown, so the rule stays silent instead of guessing. ---

func reassigned(err error, alt string) error {
	msg := err.Error()
	msg = alt

	return errors.New(msg)
}

// --- Non-applicability: a constructor from a package the rule does not know. ---

func foreignCtor(err error) error {
	return gderror.New(err.Error())
}
