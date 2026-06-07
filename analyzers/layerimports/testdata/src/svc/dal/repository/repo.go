// Eval: repository работает только с entity.
package repository

import (
	"svc/client/billing" // want `GID-228: package "svc/dal/repository" must not import "svc/client/billing"\. Fix: dal does not call external APIs directly; the client is wired in app`
	"svc/dal/entity"

	"svc/domain/model" // want `GID-132: package "svc/dal/repository" must not import "svc/domain/model"\. Fix: the dal layer works only with entity, domain types are not available to it`
)

type Snapshot struct{}

// Негатив: entity в repo — норма.
func (s *Snapshot) Snapshot(id string) (entity.Snapshot, error) {
	return entity.Snapshot{ID: id}, nil
}

// Позитив выше: model в dal-слое запрещён.
func (s *Snapshot) leak(in *model.Snapshot) {}

// Позитив выше (GID-228): внешние API дёргает client, его wiring'ует app.
func (s *Snapshot) leakClient(c *billing.Client) {}
