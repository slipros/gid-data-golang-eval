// Eval of GID-242 with a CUSTOM settings.packages list: a project uses its
// own errors facade "myerrors" (not stdlib "errors", not github.com/pkg/errors).
// Under the DEFAULT whitelist this file is clean; TestCustomPackages runs the
// analyzer with Settings{Packages: ["myerrors"]}, under which the mapper is
// flagged and the bool-predicate stays legitimate.
package customfacade

import "myerrors"

// ErrX is a sentinel used below.
var ErrX = myerrors.New("x")

// --- Positive (only under settings.packages=["myerrors"]): a mapper that
// classifies via the facade's Is and returns error. ---

func mapWithFacade(err error) error { // want `GID-242: a dedicated error-mapper function is forbidden`
	if myerrors.Is(err, ErrX) {
		return myerrors.New("mapped")
	}
	return err
}

// --- Negative: a bool-predicate via the facade — returns bool, not a mapper. ---

func isFacadeErr(err error) bool {
	return myerrors.Is(err, ErrX)
}
