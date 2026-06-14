// Negative (GID-170): event depends on domain/model and converts
// model <-> DTO — importing model in the event layer is allowed.
package producer

import (
	"svc/domain/model"
	"svc/event/dto"
)

type Snapshot struct{}

func (p *Snapshot) Publish(in model.Snapshot) dto.SnapshotDTO {
	return dto.SnapshotDTO{ID: in.ID}
}
