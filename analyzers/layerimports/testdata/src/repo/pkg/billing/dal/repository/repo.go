// GID-132 inside pkg/<module>: repository works only with entity, domain
// types are not available to it even inside an application module.
package repository

import (
	"repo/pkg/billing/dal/entity"

	"repo/pkg/billing/domain/model" // want `GID-132: package "repo/pkg/billing/dal/repository" must not import "repo/pkg/billing/domain/model"\. Fix: the dal layer works only with entity, domain types are not available to it`
)

// Snapshot is the billing repository.
type Snapshot struct{}

// Fetch — negative: entity in repo is fine.
func (s *Snapshot) Fetch(id string) (entity.Invoice, error) {
	return entity.Invoice{ID: id}, nil
}

// leak — positive above: model is forbidden in the dal layer.
func (s *Snapshot) leak(in *model.Invoice) {}
