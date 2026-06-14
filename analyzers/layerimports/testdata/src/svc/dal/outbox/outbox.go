// Positive (GID-170): dal does not depend on the event layer.
package outbox

import (
	"svc/dal/entity"
	"svc/event/dto" // want `GID-170: package "svc/dal/outbox" must not import "svc/event/dto"\. Fix: dal does not depend on the event layer; event converts model <-> DTO, not the other way`
)

type Snapshot struct{}

// Negative: entity in dal is fine.
func (o *Snapshot) Store(in entity.Snapshot) {}

// Positive above: event DTO is forbidden in the dal layer.
func (o *Snapshot) leak(in dto.SnapshotDTO) {}
