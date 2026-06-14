// Negative (GID-224): domain/model is the only layer available to
// transport; importing model in handler is allowed.
package handler

import "svc/domain/model"

type Snapshot struct{}

func (h *Snapshot) Get() model.Snapshot {
	return model.Snapshot{}
}
