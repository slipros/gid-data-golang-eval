// Eval of settings.layers: this project judges its transport too, so the core
// dal import here is reported — with the default ["domain","dal"] it would not be.
package http

import (
	entity "custom/internal/dal/entity" // want `GID-259: module package "custom/pkg/billing/server/http" must not import the core layer "custom/internal/dal/entity"`
)

// Handler — the module's transport.
type Handler struct {
	row entity.Invoice
}
