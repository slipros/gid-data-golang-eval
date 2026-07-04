// Positive (GID-235): a convert package reaches into the neighboring
// dal/repository business layer.
package convert

import (
	"svc/dal/repository" // want `GID-235: convert package "svc/dal/repository/convert" must not import "svc/dal/repository" — a converter is a pure function over vocabulary types \(model/entity/dto/client/pb\); business logic and side effects live in their layers`
)

func FromRepository(r *repository.Snapshot) string {
	return r.ID
}
