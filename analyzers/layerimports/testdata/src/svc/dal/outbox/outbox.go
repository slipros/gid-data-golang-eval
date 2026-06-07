// Позитив (GID-170): dal не зависит от event-слоя.
package outbox

import (
	"svc/dal/entity"
	"svc/event/dto" // want `GID-170: package "svc/dal/outbox" must not import "svc/event/dto"\. Fix: dal does not depend on the event layer; event converts model <-> DTO, not the other way`
)

type Snapshot struct{}

// Негатив: entity в dal — норма.
func (o *Snapshot) Store(in entity.Snapshot) {}

// Позитив выше: event-DTO в dal-слое запрещён.
func (o *Snapshot) leak(in dto.SnapshotDTO) {}
