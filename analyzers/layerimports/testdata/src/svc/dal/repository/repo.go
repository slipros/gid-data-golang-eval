// Eval: repository works only with entity.
package repository

import (
	"svc/client/billing"
	"svc/dal/entity"

	"svc/domain/model" // want `GID-132: package "svc/dal/repository" must not import "svc/domain/model"\. Fix: the dal layer works only with entity, domain types are not available to it`
)

type Snapshot struct{}

// Negative: entity in repo is fine.
func (s *Snapshot) Snapshot(id string) (entity.Snapshot, error) {
	return entity.Snapshot{ID: id}, nil
}

// Positive above: model is forbidden in the dal layer.
func (s *Snapshot) leak(in *model.Snapshot) {}

// Negative (GID-228/GID-267): the repository is the one layer allowed to call a
// client — it converts the client's models to entity in dal/repository/convert.
func (s *Snapshot) leakClient(c *billing.Client) {}
