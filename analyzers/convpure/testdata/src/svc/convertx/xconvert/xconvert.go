// Boundary (GID-235): the package path ends with "xconvert", not the exact
// "convert" segment — out of scope even though it imports a business layer.
package xconvert

import "svc/domain/service"

func Leak(s *service.Snapshot) string {
	return s.ID
}
