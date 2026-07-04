// Negative (GID-235): domain/model, dal/entity, client/*, event/dto and the
// stdlib are all vocabulary/allowed — a convert package may freely depend on
// them.
package convert

import (
	"time"

	"svc/client/billing"
	"svc/dal/entity"
	"svc/domain/model"
	"svc/event/dto"
)

type Snapshot struct {
	Model    model.Snapshot
	Entity   entity.Snapshot
	Client   billing.Client
	DTO      dto.SnapshotDTO
	MappedAt time.Time
}

func FromEntity(in entity.Snapshot, at time.Time) Snapshot {
	return Snapshot{Entity: in, MappedAt: at}
}
