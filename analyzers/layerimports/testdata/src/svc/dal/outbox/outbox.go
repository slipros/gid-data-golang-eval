// Позитив (GID-170): dal не зависит от event-слоя.
package outbox

import (
	"svc/dal/entity"
	"svc/event/dto" // want `GID-170: пакету "svc/dal/outbox" запрещён импорт "svc/event/dto" — dal не зависит от event-слоя: event конвертирует model <-> DTO, не наоборот`
)

type Snapshot struct{}

// Негатив: entity в dal — норма.
func (o *Snapshot) Store(in entity.Snapshot) {}

// Позитив выше: event-DTO в dal-слое запрещён.
func (o *Snapshot) leak(in dto.SnapshotDTO) {}
