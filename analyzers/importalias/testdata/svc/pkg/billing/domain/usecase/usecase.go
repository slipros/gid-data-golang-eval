// Boundary (GID-275): inside pkg/<module>, the shared internal/** import keeps
// the alias GID-240 demands — and holds only that alias, so the module's own
// service does not need one; a common prefix on the module's own package marks
// nothing borrowed.
package usecase

import (
	svc "example.com/svc/internal/domain/model" // GID-240 judges this alias, not GID-275
	commonservice "example.com/svc/internal/domain/service"
	commonmodel "example.com/svc/pkg/billing/domain/model"      // want `GID-275: import alias commonmodel is not needed — nothing else in the file is called model`
	billingservice "example.com/svc/pkg/billing/domain/service" // want `GID-275: import alias billingservice is not needed — nothing else in the file is called service`
)

var _ = commonservice.X + svc.X + billingservice.X + commonmodel.X
