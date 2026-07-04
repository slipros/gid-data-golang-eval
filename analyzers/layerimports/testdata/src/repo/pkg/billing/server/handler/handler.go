// GID-224 inside pkg/<module>: transport sees only domain/model of the
// service layers even inside an application module. Also proves the module
// boundary fix: repo/internal/dal/entity is a different module from
// repo/pkg/billing's point of view, so it is not flagged even though its
// path contains the "dal" segment banned for transport.
package handler

import (
	"repo/internal/dal/entity" // negative: a different module (internal/), the matrix does not apply despite the "dal" segment

	"repo/pkg/billing/domain/model"

	"repo/pkg/billing/domain/service" // want `GID-224: package "repo/pkg/billing/server/handler" must not import "repo/pkg/billing/domain/service"\. Fix: transport works only with domain/model; services and dependencies are injected as interfaces at the consumer`
)

// Snapshot is the billing transport handler.
type Snapshot struct {
	svc *service.Snapshot
}

// Get — negative: model in handler is fine (transport input/output).
func (h *Snapshot) Get() model.Invoice {
	return model.Invoice{}
}

// leak — negative: consuming a cross-module dal entity is not banned by the matrix.
func (h *Snapshot) leak(in entity.Invoice) {}
