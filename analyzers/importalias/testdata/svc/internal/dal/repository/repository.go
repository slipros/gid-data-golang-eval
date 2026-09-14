// Negative (GID-275): a package-level declaration named model makes the alias
// mandatory — an import and a package-scope object cannot share a name.
package repository

import (
	entitymodel "example.com/svc/internal/domain/model"
)

type model struct{}

var (
	_ = entitymodel.X
	_ model
)
