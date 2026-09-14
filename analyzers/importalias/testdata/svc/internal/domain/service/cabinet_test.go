// Positive (GID-275): a _test.go file is judged — nothing forces a test to
// rename a package.
package service

import (
	domainmodel "example.com/svc/internal/domain/model" // want `GID-275: import alias domainmodel is not needed`
)

var _ = domainmodel.X
