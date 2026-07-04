// Positive (GID-235): a convert package reaches into the neighboring
// dal/repository business layer.
package convert

import (
	"svc/dal/repository" // want `GID-235: convert package "svc/dal/repository/convert" must not import "svc/dal/repository" — a converter is a pure function over vocabulary types\. Fix: import only model/entity/dto/client/pb; move the logic or side effect to its layer and pass the result into the converter`
)

func FromRepository(r *repository.Snapshot) string {
	return r.ID
}
