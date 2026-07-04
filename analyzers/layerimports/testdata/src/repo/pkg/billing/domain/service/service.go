// GID-132 inside pkg/<module>: a service depends on the repository through
// an interface next to the consumer; the concrete dal/repository package is
// banned even inside an application module.
package service

import (
	"repo/internal/domain/model" // negative: a shared entity from internal/ — a different module, the matrix does not apply

	"repo/pkg/billing/dal/repository" // want `GID-132: package "repo/pkg/billing/domain/service" must not import "repo/pkg/billing/dal/repository"\. Fix: a service depends on the repository through an interface next to the consumer`
)

// Snapshot is the billing service.
type Snapshot struct {
	repo *repository.Snapshot
}

// fromShared — negative: common entities from internal/ may be consumed directly.
func (s *Snapshot) fromShared(in model.Shared) {}
