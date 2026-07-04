// Positive (GID-235): a convert package reaches into domain/usecase;
// boundary: event/dto is a vocabulary package and is allowed even though it
// lives under the otherwise-banned event segment (the event/dto exception).
package convert

import (
	"svc/domain/usecase" // want `GID-235: convert package "svc/event/consumer/convert" must not import "svc/domain/usecase" — a converter is a pure function over vocabulary types\. Fix: import only model/entity/dto/client/pb; move the logic or side effect to its layer and pass the result into the converter`
	"svc/event/dto"
)

func FromUpload(u *usecase.Upload) dto.SnapshotDTO {
	return dto.SnapshotDTO{ID: u.ID}
}
