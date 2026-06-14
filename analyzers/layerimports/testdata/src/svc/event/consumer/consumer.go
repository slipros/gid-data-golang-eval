// Positive (GID-224): consumer reaches into dal and domain/service — forbidden;
// negative: model and event/dto are allowed.
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

// Negative: model and DTO in the event layer are fine (model <-> DTO conversion).
func (c *Snapshot) Handle(in dto.SnapshotDTO) model.Snapshot {
	return model.Snapshot{ID: in.ID}
}

// Positive above: entity is not available to the consumer.
func (c *Snapshot) leak(in entity.Snapshot) {}
