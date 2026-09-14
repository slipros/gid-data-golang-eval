// Non-applicability (GID-275): GOPATH mode, no module to tell own packages by.
package service

import (
	domainmodel "nomod/internal/domain/model"
)

var _ = domainmodel.X
