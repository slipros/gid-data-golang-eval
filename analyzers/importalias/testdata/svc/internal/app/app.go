// Positive (GID-275): outside pkg/<module> a common prefix marks nothing —
// internal/** importing internal/** is an ordinary import.
package app

import (
	commonmodel "example.com/svc/internal/domain/model" // want `GID-275: import alias commonmodel is not needed`
)

var _ = commonmodel.X
