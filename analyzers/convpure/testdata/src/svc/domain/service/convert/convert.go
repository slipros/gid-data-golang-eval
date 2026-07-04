// Positive (GID-235): a convert package reaches into the neighboring
// domain/service business layer instead of staying a pure vocabulary mapper.
package convert

import (
	"svc/domain/service" // want `GID-235: convert package "svc/domain/service/convert" must not import "svc/domain/service" — a converter is a pure function over vocabulary types\. Fix: import only model/entity/dto/client/pb; move the logic or side effect to its layer and pass the result into the converter`
)

func FromService(s *service.Snapshot) string {
	return s.ID
}
