// Eval of GID-242 shape (c): a mapper that classifies its error parameter by
// BRANCHING on a value derived from it — a switch on the gRPC status code, a
// type switch, an == comparison — rather than through errors.Is/As or a
// bool-predicate. A local assigned from the parameter stands for it.
package svc

import (
	pkgerrors "github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrCabinetNotFound, ErrCabinetDisabled and ErrRegistryUnavailable are the
// domain errors the mappers below translate into.
var (
	ErrCabinetNotFound     = pkgerrors.New("cabinet not found")
	ErrCabinetDisabled     = pkgerrors.New("cabinet disabled")
	ErrRegistryUnavailable = pkgerrors.New("registry unavailable")
)

// Cabinet is a domain service calling the resource registry.
type Cabinet struct{}

// --- Positive: a method switching on the status code of its error parameter ---

func (c *Cabinet) classify(err error) error { // want `GID-242: a dedicated error-mapper function is forbidden`
	switch status.Code(pkgerrors.Cause(err)) {
	case codes.NotFound:
		return pkgerrors.WithStack(ErrCabinetNotFound)
	case codes.FailedPrecondition:
		return pkgerrors.WithStack(ErrCabinetDisabled)
	default:
		return pkgerrors.WithStack(ErrRegistryUnavailable)
	}
}

// --- Positive: a type switch on the cause of the parameter ---

func mapByTypeSwitch(err error) error { // want `GID-242: a dedicated error-mapper function is forbidden`
	switch pkgerrors.Cause(err).(type) {
	case *CustomErr:
		return ErrCabinetNotFound
	}
	return err
}

// --- Positive: an == comparison of the cause against a sentinel ---

func mapByComparison(err error) error { // want `GID-242: a dedicated error-mapper function is forbidden`
	if pkgerrors.Cause(err) == ErrX {
		return ErrCabinetNotFound
	}
	return err
}

// --- Positive: a local derived from the parameter, compared in the if ---

func mapGRPCPermissionDenied(err error) error { // want `GID-242: a dedicated error-mapper function is forbidden`
	if st, ok := status.FromError(err); ok && st.Code() == codes.PermissionDenied {
		return ErrCabinetDisabled
	}
	return err
}

// --- Positive: a chain of derived locals, the last one declared with var ---

func mapDerivedChain(err error) error { // want `GID-242: a dedicated error-mapper function is forbidden`
	cause := pkgerrors.Cause(err)
	var st, _ = status.FromError(cause)
	switch st.Code() {
	case codes.NotFound:
		return ErrCabinetNotFound
	}
	return err
}

// --- Positive: shapes (a) and (b) see a derived local as the parameter too ---

func mapCauseIs(err error) error { // want `GID-242: a dedicated error-mapper function is forbidden`
	cause := pkgerrors.Cause(err)
	if pkgerrors.Is(cause, ErrX) {
		return ErrCabinetNotFound
	}
	return err
}

func isUnavailable(err error) bool { return err == ErrRegistryUnavailable }

func mapCausePredicate(err error) error { // want `GID-242: a dedicated error-mapper function is forbidden`
	cause := pkgerrors.Cause(err)
	if isUnavailable(cause) {
		return ErrCabinetDisabled
	}
	return err
}

// --- Negative: a nil check asks whether there is an error, it classifies nothing ---

func wrapUnlessNil(err error) error {
	if err == nil {
		return nil
	}
	if nil != err {
		return pkgerrors.WithStack(err)
	}
	return err
}

// --- Negative: a bool-predicate branching on the status code ---

func isNotFoundCode(err error) bool {
	return status.Code(err) == codes.NotFound
}

// --- Negative: an observer switches on the code only to count, and hands err back ---

func countByCode(err error) error {
	switch status.Code(err) {
	case codes.NotFound:
		notFoundTotal++
	}
	return err
}

var notFoundTotal int

// --- Boundary: the branch is on a local that is not derived from the parameter ---

func (c *Cabinet) fetch() error { return nil }

func (c *Cabinet) retryOnce(err error) error {
	res := c.fetch()
	if res == ErrX {
		return ErrRegistryUnavailable
	}
	return pkgerrors.WithStack(err)
}

// --- Boundary: a tagless switch whose cases compare nothing derived from err ---

func (c *Cabinet) pick(err error, disabled bool) error {
	switch {
	case disabled:
		return ErrCabinetDisabled
	}
	return pkgerrors.Wrap(err, "pick")
}
