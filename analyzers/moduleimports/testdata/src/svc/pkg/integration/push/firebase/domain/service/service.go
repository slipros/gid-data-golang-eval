// Eval for GID-259 (module-core-imports): the module's own service layer,
// nested two directories deeper than pkg/<module> — the layout that made every
// layer rule silent before the pathseg fix (incident 2026-08-06).
package service

import (
	"context"

	// Positive: the core data layer — the module owns a dal of its own.
	commonentity "svc/internal/dal/entity" // want `GID-259: module package "svc/pkg/integration/push/firebase/domain/service" must not import the core layer "svc/internal/dal/entity" — a module owns its dal, and only the core /domain/\*\* is shared\. Fix: declare the repository interface over the module's own entity \(<module>/dal/entity\), and take core data through a core service injected in module\.go`

	// Positive: the core repository is wiring material, not a dependency.
	commonrepository "svc/internal/dal/repository" // want `GID-259: module package "svc/pkg/integration/push/firebase/domain/service" must not import the core layer "svc/internal/dal/repository"`

	// Positive: a core client is not a shared layer either.
	commonclient "svc/internal/client/vendor" // want `GID-259: module package "svc/pkg/integration/push/firebase/domain/service" must not import the core layer "svc/internal/client/vendor"`

	// Negative: the shared vocabulary and the core business layer are allowed.
	commonmodel "svc/internal/domain/model"
	commonservice "svc/internal/domain/service"

	// Negative: the module's own layers are its own business.
	"svc/pkg/integration/push/firebase/dal/entity"
)

// CoreIntegration — the core service the module takes core data through.
type CoreIntegration interface {
	Integration(ctx context.Context, id string) (string, error)
}

// IntegrationRepository — the repository over the module's own entity.
type IntegrationRepository interface {
	Integration(ctx context.Context, id string) (entity.Integration, error)
}

// Integration — the module's service.
type Integration struct {
	core CoreIntegration
	repo IntegrationRepository
}

// Core returns the core record through the core service.
func (i *Integration) Core(ctx context.Context, id string) (commonmodel.Integration, error) {
	if _, err := i.core.Integration(ctx, id); err != nil {
		return commonmodel.Integration{}, err
	}

	return commonmodel.Integration{ID: id}, nil
}

// coreEntity, coreRepo and coreClient keep the forbidden imports used.
var (
	_ = commonentity.Integration{}
	_ = commonrepository.Integration{}
	_ = commonclient.Client{}
	_ = commonservice.Integration{}
)
