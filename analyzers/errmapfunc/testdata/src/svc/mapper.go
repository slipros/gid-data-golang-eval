// Eval of GID-242: a dedicated error-MAPPER function — one that classifies
// its own error parameter via errors.Is/errors.As AND returns error (maps
// error to error/status) — is forbidden. The bounded set of errors must be
// mapped inline, at the call site (handler/interceptor). This is NOT specific
// to gRPC: any error return counts. The RETURN type is a discriminator — a
// bool-predicate (isNotFound/isRetryable/isCustom) classifies but does not
// map, and is legitimate.
package svc

import (
	"errors"
	"fmt"
	"net/http"

	pkgerrors "github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrX is a sentinel error used across the scenarios below.
var ErrX = errors.New("x")

// CustomErr is a typed error used for the errors.As scenarios.
type CustomErr struct {
	Msg string
}

func (e *CustomErr) Error() string { return e.Msg }

// --- Positive: a mapper returning error (gRPC status), via errors.Is ---

func mapErr(err error) error { // want `GID-242: a dedicated error-mapper function is forbidden`
	switch {
	case errors.Is(err, ErrX):
		return status.Error(codes.NotFound, "not found")
	default:
		return status.Error(codes.Internal, "internal error")
	}
}

// --- Positive: a mapper via github.com/pkg/errors.Is (the gid.team default,
// GID-146) — its package path is github.com/pkg/errors, not "errors", but it
// is the same classification API and must be flagged. This is the real-code
// case the stdlib-only whitelist was missing. ---

func mapPkgErr(err error) error { // want `GID-242: a dedicated error-mapper function is forbidden`
	if pkgerrors.Is(err, ErrX) {
		return status.Error(codes.NotFound, "not found")
	}
	return err
}

// --- Positive: a mapper via github.com/pkg/errors.As ---

func mapPkgErrAs(err error) error { // want `GID-242: a dedicated error-mapper function is forbidden`
	var t *CustomErr
	if pkgerrors.As(err, &t) {
		return status.Error(codes.Internal, t.Msg)
	}
	return err
}

// --- Positive: a mapper classifying via errors.As (type-assert) and returning error ---

func mapErrAs(err error) error { // want `GID-242: a dedicated error-mapper function is forbidden`
	var t *CustomErr
	if errors.As(err, &t) {
		return status.Error(codes.Internal, t.Msg)
	}
	return err
}

// --- Positive: a mapper with a (T, error) result — the error result still makes it a mapper ---

func mapErrTuple(err error) (int, error) { // want `GID-242: a dedicated error-mapper function is forbidden`
	if errors.Is(err, ErrX) {
		return 0, status.Error(codes.NotFound, "not found")
	}
	return 0, nil
}

// --- Negative: a bool-predicate classifies the error (errors.Is) but does not map it ---

func isRetryable(err error) bool {
	return errors.Is(err, ErrX)
}

func isNotFound(err error) bool {
	switch {
	case errors.Is(err, ErrX):
		return true
	default:
		return false
	}
}

// --- Negative: a bool-predicate classifies via errors.As but does not map it ---

func isCustom(err error) bool {
	var t *CustomErr
	return errors.As(err, &t)
}

// --- Negative: a bool-predicate via github.com/pkg/errors.Is — the return-type
// discriminator holds regardless of which errors package is used ---

func isPkgRetryable(err error) bool {
	return pkgerrors.Is(err, ErrX)
}

// --- Negative: classifies via errors.Is but returns a plain int (HTTP status code),
// not an error — by the return-type discriminator this is not a mapper. ---

func mapToHTTPStatus(err error) int {
	switch {
	case errors.Is(err, ErrX):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

// --- Negative: inline handling in a handler — errors.Is branches on a LOCAL
// variable (the usecase call result), not on a function parameter. ---

// UseCase is the injected dependency called by Handler.
type UseCase interface {
	Do() (int, error)
}

// Handler demonstrates the legitimate inline shape.
type Handler struct {
	u UseCase
}

func (h *Handler) Handle() (int, error) {
	res, err := h.u.Do()
	if err != nil {
		switch {
		case errors.Is(err, ErrX):
			return 0, status.Error(codes.NotFound, "not found")
		}
	}
	return res, nil
}

// --- Negative: returns error but never calls errors.Is/errors.As (a plain wrapper) ---

func wrap(err error) error {
	return fmt.Errorf("wrap: %w", err)
}

// --- Negative: no error parameter — a plain request validator ---

// Req is a request struct validated below.
type Req struct {
	Name string
}

func validate(req Req) error {
	if req.Name == "" {
		return status.Error(codes.InvalidArgument, "name is required")
	}
	return nil
}

// --- Boundary: an error parameter, returns error, but no errors.Is/errors.As on it ---

func passthrough(err error) error {
	return err
}

// --- Non-applicability: an unnamed error parameter cannot be referenced by errors.Is/As ---

func discard(error) error {
	return status.Error(codes.Internal, "x")
}
