// Negative: chain inspection through pkg/errors — the whole file needs one
// errors package, so nothing is reported.
package onlypkgerrors

import "github.com/pkg/errors"

var errNoResult = errors.New("no result")

func cleanIsNotFound(err error) bool {
	return errors.Is(err, errNoResult)
}

func cleanAs(err error, target any) bool {
	return errors.As(err, target)
}

func cleanUnwrap(err error) error {
	return errors.Unwrap(err)
}
