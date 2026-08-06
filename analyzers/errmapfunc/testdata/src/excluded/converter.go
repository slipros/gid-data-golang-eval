// Eval of GID-242 settings.exclude: some frameworks DICTATE the signature of
// the one place error translation is allowed to live — see the package doc
// for gdgrpcserver.WithErrorConverters (interceptor.ErrorConverterFunc =
// func(error) *status.Status). ValidationErrorConverter below has the exact
// mapper shape the rule forbids (classifies its error parameter via
// errors.As, hands back *status.Status) and is cleared only because it is on
// the exclusion list ("ValidationErrorConverter"). Converter has the
// identical shape but is NOT on the list — it must still be reported, which
// is what proves the setting drives the exclusion rather than a blanket
// exemption for the shape.
package excluded

import (
	pkgerrors "github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ValidationErr is a typed error classified via errors.As, standing in for
// the real validator.Result classified in production.
type ValidationErr struct {
	Field string
}

func (e *ValidationErr) Error() string { return e.Field }

// ValidationErrorConverter mirrors the canonical resource-registry /
// advertising-api converter registered in gdgrpcserver.WithErrorConverters.
// Excluded by name — no diagnostic.
func ValidationErrorConverter(err error) *status.Status {
	var t *ValidationErr
	if !pkgerrors.As(err, &t) {
		return nil
	}
	return status.New(codes.InvalidArgument, t.Field)
}

// Converter has the identical mapper shape as ValidationErrorConverter but is
// not on settings.exclude — still a mapper.
func Converter(err error) *status.Status { // want `GID-242: a dedicated error-mapper function is forbidden`
	var t *ValidationErr
	if !pkgerrors.As(err, &t) {
		return nil
	}
	return status.New(codes.Internal, t.Field)
}
