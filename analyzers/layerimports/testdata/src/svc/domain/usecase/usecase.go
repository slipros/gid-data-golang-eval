// Eval: usecase works only with model.
package usecase

import (
	"svc/dal/entity"     // want `GID-132: package "svc/domain/usecase" must not import "svc/dal/entity"\. Fix: usecase works only with model and talks to DAL through services`
	"svc/dal/repository" // want `GID-132: package "svc/domain/usecase" must not import "svc/dal/repository"\. Fix: usecase works only with model and talks to DAL through services`

	"svc/domain/model"
	"svc/domain/model/filter"
)

type Upload struct {
	repo *repository.Snapshot
}

func (u *Upload) bad(id string) (entity.Snapshot, error) {
	return u.repo.Snapshot(id)
}

// Negative: model in usecase is fine.
func (u *Upload) good() model.Snapshot {
	return model.Snapshot{}
}

// Negative (boundary): the nested packages /domain/model/* are a full-fledged
// model layer, usecase accepts and returns their types.
func (u *Upload) goodFilter(f *filter.Snapshots) []string {
	return f.IDs
}
