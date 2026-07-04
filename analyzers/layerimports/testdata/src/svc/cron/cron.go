// Eval for GID-241: a folder the deny matrix knows nothing about still must
// not import the repository — the allow-list (app + the repository layer)
// covers future layout drift by default.
package cron

import (
	"svc/dal/repository" // want `GID-241: package "svc/cron" must not import "svc/dal/repository" — a repository is wired in app and consumed by services through an interface \(GID-132/134\)\. Fix: declare an <Entity>Repository interface next to the consumer and inject the concrete repository in the composition root`
	"svc/domain/model"
)

// Tick uses the repository directly — a violation regardless of the folder name.
func Tick(r *repository.Snapshot) model.Snapshot {
	_ = r
	return model.Snapshot{}
}
