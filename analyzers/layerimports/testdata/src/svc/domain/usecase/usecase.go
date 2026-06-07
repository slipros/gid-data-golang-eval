// Eval: usecase работает только с model.
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

// Негатив: model в usecase — норма.
func (u *Upload) good() model.Snapshot {
	return model.Snapshot{}
}

// Негатив (граница): вложенные пакеты /domain/model/* — полноправный
// model-слой, usecase принимает и возвращает их типы.
func (u *Upload) goodFilter(f *filter.Snapshots) []string {
	return f.IDs
}
