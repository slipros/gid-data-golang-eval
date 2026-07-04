// Eval for GID-224: a background job is a leaf like transport — it sees only
// domain/model; repositories and services arrive as interfaces at the consumer.
package refresh

import (
	"svc/dal/repository" // want `GID-224: package "svc/job/refresh" must not import "svc/dal/repository"\. Fix: a background job works only with domain/model; services and dependencies are injected as interfaces at the consumer`
	"svc/domain/model"
)

// Job refreshes snapshots in the background.
type Job struct {
	repo *repository.Snapshot
}

// Snapshot returns a vocabulary value: importing domain/model is allowed.
func (j *Job) Snapshot() model.Snapshot {
	_ = j.repo
	return model.Snapshot{}
}
