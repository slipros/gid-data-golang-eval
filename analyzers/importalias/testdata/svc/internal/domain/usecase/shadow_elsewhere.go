// Boundary (GID-275): the shadowing parameter lives in a function that does
// not use the import — it justifies nothing.
package usecase

import (
	usecasemodel "example.com/svc/internal/domain/model" // want `GID-275: import alias usecasemodel is not needed — nothing else in the file is called model`
)

func Other() int {
	return usecasemodel.X
}

func Third(model int) int {
	return model
}
