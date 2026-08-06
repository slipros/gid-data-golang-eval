// Non-applicability: the module root is its composition root — wiring is where
// naming the concrete core repository is legitimate (the same carve-out GID-241
// makes for /app/**). No diagnostic here.
package firebase

import (
	commonrepository "svc/internal/dal/repository"
	commonservice "svc/internal/domain/service"

	domainservice "svc/pkg/integration/push/firebase/domain/service"
)

// Module — the module wired from the core dependencies.
type Module struct {
	service *domainservice.Integration
}

// NewModule wires the module.
func NewModule() *Module {
	_ = commonrepository.Integration{}
	_ = commonservice.Integration{}

	return &Module{}
}
