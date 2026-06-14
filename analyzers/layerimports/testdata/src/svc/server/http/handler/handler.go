// Positive (GID-224): server imports dal and domain/service — forbidden;
// negative: domain/model and validate are allowed for transport.
package handler

import (
	"svc/dal/repository" // want `GID-224: package "svc/server/http/handler" must not import "svc/dal/repository"\. Fix: transport works only with domain/model; services and dependencies are injected as interfaces at the consumer`
	"svc/domain/model"
	"svc/domain/service" // want `GID-224: package "svc/server/http/handler" must not import "svc/domain/service"\. Fix: transport works only with domain/model; services and dependencies are injected as interfaces at the consumer`
	"svc/validate"
)

type Snapshot struct {
	svc  *service.Snapshot
	repo *repository.Snapshot
	v    *validate.Snapshot
}

// Negative: model in handler is fine (transport input/output).
func (h *Snapshot) Get() model.Snapshot {
	return model.Snapshot{}
}
