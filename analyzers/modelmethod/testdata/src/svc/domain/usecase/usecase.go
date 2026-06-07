// Eval для GID-195: usecase — тоже scope правила.
package usecase

import "svc/domain/model"

func snapshotKey(s *model.Snapshot) string { // want `GID-195: приватная функция "snapshotKey" работает только со значением model.Snapshot — это поведение модели: оформите её публичным методом этого типа`
	return s.ID
}
