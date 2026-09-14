// Negative (GID-275): a nested module borrows its parent module's model under
// a common prefix — the same own-versus-borrowed marker GID-240 sets for
// internal/**.
package service

import (
	commonmodel "example.com/svc/pkg/integration/domain/model"
)

var _ = commonmodel.X
