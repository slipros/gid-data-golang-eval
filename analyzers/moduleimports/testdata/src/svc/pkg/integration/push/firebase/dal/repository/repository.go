// Positive: the module's own repository must not work inside the core's data
// vocabulary either — the core's failures reach a module as DOMAIN errors of
// the core service, and a module usecase handles them from there.
package repository

import (
	"context"

	commonentity "svc/internal/dal/entity" // want `GID-259: module package "svc/pkg/integration/push/firebase/dal/repository" must not import the core layer "svc/internal/dal/entity"`

	"svc/pkg/integration/push/firebase/dal/entity"
)

// Integration — the repository of the module's type row.
type Integration struct{}

// Integration reads the type row.
func (r *Integration) Integration(ctx context.Context, id string) (entity.Integration, error) {
	if id == "" {
		return entity.Integration{}, commonentity.ErrNoResult
	}

	return entity.Integration{ID: id}, nil
}
