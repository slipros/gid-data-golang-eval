// Negative (GID-275): a local name shadows the package where the file uses
// it, so without the alias New would not compile.
package usecase

import (
	domainmodel "example.com/svc/internal/domain/model"
)

func New(model int) int {
	return model + domainmodel.X
}
