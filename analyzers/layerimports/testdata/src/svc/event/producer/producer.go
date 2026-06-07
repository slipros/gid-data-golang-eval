// Негатив (GID-170): event зависит от domain/model и конвертирует
// model <-> DTO — импорт model в event-слое разрешён.
package producer

import (
	"svc/domain/model"
	"svc/event/dto"
)

type Snapshot struct{}

func (p *Snapshot) Publish(in model.Snapshot) dto.SnapshotDTO {
	return dto.SnapshotDTO{ID: in.ID}
}
