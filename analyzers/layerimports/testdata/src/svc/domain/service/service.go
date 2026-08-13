// Eval: service converts model <-> entity, but depends on the repository
// through an interface, not by importing the implementation. It reaches an
// external client the same way — through a repository (GID-267).
package service

import (
	"svc/client/billing"             // want `GID-267: package "svc/domain/service" must not import "svc/client/billing"\. Fix: call the client from dal/repository and convert its models to entity in dal/repository/convert, then depend on that repository through an interface next to the service`
	"svc/connect/client/interceptor" // boundary: a nested client segment is not the client layer
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

// Positive (GID-267): a service must not reach a client directly — the chain is
// client -> repository -> service. Positive above: metrics — through an
// interface (GID-226).
func (s *Snapshot) leakDeps(c *billing.Client, m *metric.Prometheus) {}

// Negative (boundary, GID-267): connect/client/interceptor only carries a
// "client" segment nested below another layer — the layer is anchored to the
// first segment after the module root ("connect"), so this is not the client
// layer and the rule does not fire.
func (s *Snapshot) wrap(in model.Snapshot) model.Snapshot {
	return interceptor.Wrap(in)
}
