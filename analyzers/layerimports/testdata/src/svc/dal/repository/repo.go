// Eval: repository works only with entity.
package repository

import (
	"svc/client/billing" // want `GID-228: package "svc/dal/repository" must not import "svc/client/billing"\. Fix: dal does not call external APIs directly; the client is wired in app`
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

// Positive above (GID-228): external APIs are called by the client, which app wires.
func (s *Snapshot) leakClient(c *billing.Client) {}
