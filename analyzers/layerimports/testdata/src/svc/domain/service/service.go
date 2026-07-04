// Eval: service converts model <-> entity, but depends on the repository
// through an interface, not by importing the implementation.
package service

import (
	"svc/client/billing"
	"svc/dal/entity"
	"svc/dal/repository" // want `GID-132: package "svc/domain/service" must not import "svc/dal/repository"\. Fix: a service depends on the repository through an interface next to the consumer`
	"svc/metric"         // want `GID-226: package "svc/domain/service" must not import "svc/metric"\. Fix: domain receives metrics through an interface; the metric package is wired in app`

	"svc/domain/model"
)

// Negative (boundary): importing entity is allowed for a service — conversion.
type Snapshot struct {
	repo *repository.Snapshot
}

func (s *Snapshot) Snapshot(id string) (model.Snapshot, error) {
	out, err := s.repo.Snapshot(id)
	if err != nil {
		return model.Snapshot{}, err
	}
	return fromEntity(&out), nil
}

func fromEntity(in *entity.Snapshot) model.Snapshot {
	return model.Snapshot{ID: in.ID}
}

// Negative (GID-228): a service is allowed to call a client directly — its API
// always takes and returns model, conversion model <-> client models is internal.
// Positive above: metrics — through an interface (GID-226).
func (s *Snapshot) leakDeps(c *billing.Client, m *metric.Prometheus) {}
