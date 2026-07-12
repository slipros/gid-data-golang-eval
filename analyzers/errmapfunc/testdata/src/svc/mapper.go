// Eval of GID-242: a dedicated function that classifies its own error
// parameter via errors.Is is forbidden — the bounded set of errors must be
// handled inline, at the call site (in the handler/interceptor). This is
// NOT specific to gRPC: gRPC status is just one possible target below,
// alongside a plain HTTP status code — the rule flags the SHAPE (a function
// deciding something by testing its own error parameter), not the target.
package svc

import (
	"errors"
	"fmt"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrX is a sentinel error used across the scenarios below.
var ErrX = errors.New("x")

// --- Positive: a dedicated mapper to a plain int (HTTP status) — no gRPC involved at all ---

func mapToHTTPStatus(err error) int { // want `GID-242: a dedicated function that classifies its own error parameter via errors\.Is is forbidden`
	switch {
	case errors.Is(err, ErrX):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

// --- Positive: gRPC status is just ANOTHER instance of the same forbidden shape ---

func mapErr(err error) error { // want `GID-242: a dedicated function that classifies its own error parameter via errors\.Is is forbidden`
	switch {
	case errors.Is(err, ErrX):
		return status.Error(codes.NotFound, "not found")
	default:
		return status.Error(codes.Internal, "internal error")
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

// --- Negative: not a mapper — no errors.Is at all ---

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

// --- Boundary: an error parameter, but no errors.Is on it at all ---

func passthrough(err error) error {
	return err
}

// --- Non-applicability: an unnamed error parameter cannot be referenced by errors.Is ---

func discard(error) error {
	return status.Error(codes.Internal, "x")
}
