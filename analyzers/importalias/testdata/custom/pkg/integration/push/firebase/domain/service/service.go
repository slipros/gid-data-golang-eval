// Boundary (GID-275, settings.prefix: shared): the borrowed marker is now
// shared — a common-prefixed alias marks nothing.
package service

import (
	commonmodel "example.com/custom/pkg/integration/domain/model" // want `GID-275: import alias commonmodel is not needed`
)

var _ = commonmodel.X
