// Eval of GID-242: a dedicated error-MAPPER function — one that classifies
// its own error parameter (errors.Is/errors.As, or ANY bool-predicate over an
// error such as a driver's IsNoResult) AND returns an error of its own (maps
// error to error/status) — is forbidden. The bounded set of errors must be
// mapped inline, at the call site (repository method/handler). This is NOT
// specific to gRPC: any error return counts. Two discriminators keep the
// legitimate shapes clean: the RETURN type — a bool-predicate
// (isNotFound/isRetryable/isCustom) classifies but does not map; and
// PRODUCING an error — an observer that classifies only to log and hands the
// same value back maps nothing.
package svc

import (
	"driver"
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

// --- Positive (shape b): the real-code incident — a repository-level mapper
// built on a DRIVER's bool-predicates. It never calls errors.Is/errors.As, so
// shape (a) alone was blind to it (resource-registry repository.MapError,
// 2026-08-04). It classifies its own parameter and returns errors of its own. ---

func MapError(err error) error { // want `GID-242: a dedicated error-mapper function is forbidden`
	switch {
	case driver.IsUniqueViolation(err):
		return pkgerrors.WithStack(ErrX)
	case driver.IsNoResult(err):
		return pkgerrors.WithStack(ErrX)
	default:
		return err
	}
}

// --- Positive (shape b): the predicate lives in this very package — a mapper
// does not become legitimate by keeping its classifier local. ---

func mapLocalPredicate(err error) error { // want `GID-242: a dedicated error-mapper function is forbidden`
	if isRetryable(err) {
		return status.Error(codes.Unavailable, "retry")
	}
	return err
}

// --- Positive (shape b): the error is replaced by ASSIGNING to the parameter
// rather than by returning a different expression — discriminator #3 counts
// the assignment, so the sentinel-then-return mapper is still caught. ---

func mapByAssign(err error) error { // want `GID-242: a dedicated error-mapper function is forbidden`
	if driver.IsNoResult(err) {
		err = ErrX
	}
	return err
}

// --- Positive (shape b): a method is a mapper just as a function is. ---

// Repo owns a method-shaped mapper.
type Repo struct{}

func (r *Repo) mapError(err error) error { // want `GID-242: a dedicated error-mapper function is forbidden`
	if driver.IsUniqueViolation(err) {
		return ErrX
	}
	return err
}

// --- Negative (discriminator #3): classifies via a driver predicate but
// produces no error of its own — it only decides how to observe the error and
// hands the very same value back. An observer is not a mapper. ---

func logAndReturn(err error) error {
	if driver.IsTemporary(err) {
		fmt.Println("temporary")
	} else {
		fmt.Println("permanent")
	}
	return err
}

// --- Boundary (discriminator #3): the same shape via errors.Is — the
// discriminator holds for shape (a) too, not only for the new predicates. ---

func countAndReturn(err error) error {
	if errors.Is(err, ErrX) {
		fmt.Println("known")
	}
	return err
}

// --- Negative (discriminator #2 under shape b): the predicate classifies a
// LOCAL variable (the driver call result), not the function's parameter —
// this is the legitimate inline shape the rule demands. ---

// Conn is the injected dependency of the repository below.
type Conn interface {
	Select() error
}

// Repository demonstrates inline mapping at the call site.
type Repository struct {
	conn Conn
}

func (r *Repository) Get() error {
	if err := r.conn.Select(); err != nil {
		if driver.IsNoResult(err) {
			err = ErrX
		}
		return pkgerrors.Wrap(err, "select")
	}
	return nil
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

// --- Positive (discriminator #1, widened 2026-08-06): classifies via errors.Is
// and hands back an HTTP status code. The transport type of the translation is
// irrelevant — the error is still being mapped away from its origin by a
// dedicated function. ---

func mapToHTTPStatus(err error) int { // want `GID-242: a dedicated error-mapper function is forbidden`
	switch {
	case errors.Is(err, ErrX):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

// --- Positive (the 2026-08-06 incident, resource-registry
// internal/server/grpc/errors): the mapper split into a codes.Code half and a
// *status.Status half, classifying through the package's own bool-predicates.
// Neither half returns error, so both were invisible to the first cut of
// discriminator #1 — and the package doc cited the clean lint run as proof of
// legitimacy. ---

func Code(err error) codes.Code { // want `GID-242: a dedicated error-mapper function is forbidden`
	switch {
	case isNotFound(err):
		return codes.NotFound
	case isRetryable(err):
		return codes.Unavailable
	default:
		return codes.Internal
	}
}

func Converter(err error) *status.Status { // want `GID-242: a dedicated error-mapper function is forbidden`
	if isNotFound(err) {
		return status.New(codes.NotFound, err.Error())
	}
	return status.New(codes.Internal, err.Error())
}

// --- Positive (discriminator #1 + #3): a named result filled in by the
// classified branch, with a naked return — no return expression to inspect,
// the assignment to the named result is what marks the translation. ---

func mapToNamedCode(err error) (code codes.Code) { // want `GID-242: a dedicated error-mapper function is forbidden`
	code = codes.Internal
	if driver.IsNoResult(err) {
		code = codes.NotFound
	}
	return
}

// --- Positive: classifies and hands back a message string — the flattest
// translation there is, and still a mapper. ---

func errorMessage(err error) string { // want `GID-242: a dedicated error-mapper function is forbidden`
	if errors.Is(err, ErrX) {
		return "not found"
	}
	return err.Error()
}

// --- Negative (discriminator #1): a named bool result is still a predicate —
// the discriminator is the result TYPE, not whether it carries a name. ---

func isKnown(err error) (ok bool) {
	ok = errors.Is(err, ErrX)
	return
}

// --- Negative (discriminator #1): a classifier with NO results translates the
// error into nothing at all — it only decides what to log. ---

func logClassified(err error) {
	if driver.IsTemporary(err) {
		fmt.Println("temporary")
		return
	}
	fmt.Println("permanent")
}

// --- Negative (discriminator #3 under the widened #1): an observer with a
// (T, error) signature. It fills in a zero T next to the untouched err — the
// zero T must not read as a translation, or every observer would be reported. ---

func observeTuple(err error) (int, error) {
	if driver.IsTemporary(err) {
		fmt.Println("temporary")
	}
	return 0, err
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
