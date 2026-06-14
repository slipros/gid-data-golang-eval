// Boundary: *_test.go is skipped — declaring an error here does not violate GID-169.
package model

import "github.com/pkg/errors"

var errTestOnly = errors.New("test only")
