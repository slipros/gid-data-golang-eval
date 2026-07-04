// Negative (GID-241): the pkg/<module> root package is the module's
// composition root (module.md) — module.go wires the module's repositories
// into its services, like internal/app does for the service.
package billing

import (
	"repo/pkg/billing/dal/repository"
)

// Module aggregates the billing module's wiring.
type Module struct {
	repo *repository.Snapshot
}

// NewModule assembles the module's dependencies.
func NewModule() *Module {
	return &Module{repo: &repository.Snapshot{}}
}
