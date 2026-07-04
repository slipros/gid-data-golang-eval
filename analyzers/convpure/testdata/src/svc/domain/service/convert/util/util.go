// Boundary (GID-235): "convert" is a middle segment here, not the last one
// (the package path ends with "util") — out of scope even though the import
// reaches into a business layer.
package util

import "svc/dal/repository"

func Leak(r *repository.Snapshot) string {
	return r.ID
}
