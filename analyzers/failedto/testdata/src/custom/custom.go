package custom

import "github.com/pkg/errors"

// settings.prefixes = ["oops"] замещает дефолт целиком:
// дефолтный "failed to" больше не ловится, ловится только "oops".

func wrapOops(err error) error {
	return errors.Wrap(err, "oops broken") // want `GID-184: error message starts with "oops"`
}

// "failed to" не из кастомного списка — не матчится.
func wrapFailed(err error) error {
	return errors.Wrap(err, "failed to select")
}
