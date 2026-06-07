// Позитив (GID-224): consumer лезет в dal и domain/service — запрещено;
// негатив: model и event/dto разрешены.
package consumer

import (
	"svc/dal/entity"     // want `GID-224: пакету "svc/event/consumer" запрещён импорт "svc/dal/entity" — транспорт работает только с domain/model: сервисы и зависимости инжектятся интерфейсами у потребителя`
	"svc/domain/model"
	"svc/domain/service" // want `GID-224: пакету "svc/event/consumer" запрещён импорт "svc/domain/service" — транспорт работает только с domain/model: сервисы и зависимости инжектятся интерфейсами у потребителя`
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
