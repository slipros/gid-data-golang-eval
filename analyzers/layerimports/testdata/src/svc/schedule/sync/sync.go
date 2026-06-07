// Позитив (GID-224): schedule дёргает service напрямую — запрещено;
// негатив: model разрешён.
package sync

import (
	"svc/domain/model"
	"svc/domain/service" // want `GID-224: package "svc/schedule/sync" must not import "svc/domain/service"\. Fix: transport works only with domain/model; services and dependencies are injected as interfaces at the consumer`
)

type Job struct {
	svc *service.Snapshot
}

// Негатив: model в schedule — норма.
func (j *Job) Run() model.Snapshot {
	return model.Snapshot{}
}
