// Positive (GID-235): a convert package reaches into the neighboring
// domain/service business layer instead of staying a pure vocabulary mapper.
package convert

import (
	"svc/domain/service" // want `GID-235: convert package "svc/domain/service/convert" must not import "svc/domain/service" — a converter is a pure function over vocabulary types \(model/entity/dto/client/pb\); business logic and side effects live in their layers`
)

func FromService(s *service.Snapshot) string {
	return s.ID
}
