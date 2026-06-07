// Позитив (GID-224): consumer лезет в dal и domain/service — запрещено;
// негатив: model и event/dto разрешены.
package consumer

import (
	"svc/dal/entity"     // want `GID-224: package "svc/event/consumer" must not import "svc/dal/entity"\. Fix: transport works only with domain/model; services and dependencies are injected as interfaces at the consumer`
	"svc/domain/model"
	"svc/domain/service" // want `GID-224: package "svc/event/consumer" must not import "svc/domain/service"\. Fix: transport works only with domain/model; services and dependencies are injected as interfaces at the consumer`
	"svc/event/dto"
)

type Snapshot struct {
	svc *service.Snapshot
}

// Негатив: model и DTO в event-слое — норма (конвертация model <-> DTO).
func (c *Snapshot) Handle(in dto.SnapshotDTO) model.Snapshot {
	return model.Snapshot{ID: in.ID}
}

// Позитив выше: entity консьюмеру недоступен.
func (c *Snapshot) leak(in entity.Snapshot) {}
