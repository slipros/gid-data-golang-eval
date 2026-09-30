// Inapplicable: a package outside the domain tree — the rule does not apply.
package util

import "errors"

var ErrUtil = errors.New("util")

// Non-applicability: /pkg is outside the domain tree — an error type is allowed here.
type UtilError struct{}

func (UtilError) Error() string { return "util" }
