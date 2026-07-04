// Eval for GID-224: a background job executes business logic — importing
// domain/model, domain/service and domain/usecase is allowed; infrastructure
// (dal, client, metric, event, transport, app) is not.
package refresh

import (
	"svc/dal/repository" // want `GID-224: package "svc/job/refresh" must not import "svc/dal/repository"\. Fix: a background job works through the business layers \(model/service/usecase\); infrastructure \(dal, client, transport\) is not available to it directly`
	"svc/domain/model"
	"svc/domain/service"
	"svc/domain/usecase"
)

// Job refreshes snapshots in the background.
type Job struct {
	repo *repository.Snapshot
	svc  *service.Snapshot
	uc   *usecase.Upload
}

// Snapshot returns a vocabulary value: business layers are allowed.
func (j *Job) Snapshot() model.Snapshot {
	_, _ = j.svc, j.uc
	_ = j.repo
	return model.Snapshot{}
}
