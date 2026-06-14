package custom

import "github.com/pkg/errors"

// settings.prefixes = ["oops"] replaces the default entirely:
// the default "failed to" is no longer caught, only "oops" is.

func wrapOops(err error) error {
	return errors.Wrap(err, "oops broken") // want `GID-184: error message starts with "oops"`
}

// "failed to" is not in the custom list — not matched.
func wrapFailed(err error) error {
	return errors.Wrap(err, "failed to select")
}
