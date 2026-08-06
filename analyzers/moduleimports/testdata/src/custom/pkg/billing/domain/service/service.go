// Eval of the GID-259 settings: settings.allow adds the "client" layer to the
// shared core, settings.exclude clears the core repository path. The core dal
// entity stays forbidden — the settings drive the exclusion, they are not a
// blanket exemption.
package service

import (
	vendor "custom/internal/client/vendor"
	entity "custom/internal/dal/entity" // want `GID-259: module package "custom/pkg/billing/domain/service" must not import the core layer "custom/internal/dal/entity"`
	repository "custom/internal/dal/repository"
)

// Invoice — the module's service.
type Invoice struct {
	row  entity.Invoice
	repo repository.Invoice
	cli  vendor.Client
}
