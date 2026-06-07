// Позитив (GID-170): domain не зависит от event-слоя.
package notifier

import (
	"svc/domain/model"
	"svc/event/dto" // want `GID-170: пакету "svc/domain/notifier" запрещён импорт "svc/event/dto" — domain не зависит от event-слоя: event конвертирует model <-> DTO, не наоборот`
)

type Snapshot struct{}

// Негатив: model в domain — норма.
func (n *Snapshot) Build() model.Snapshot {
	return model.Snapshot{}
}

// Позитив выше: event-DTO в domain-слое запрещён.
func (n *Snapshot) leak(in dto.SnapshotDTO) {}
