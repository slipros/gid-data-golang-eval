// Класс 4 (неприменимость) — _test.go пропускается: инлайн entity-литерал
// в тесте не флагается.
package service

import "svc/dal/entity"

func buildTestSnapshot(id string) entity.Snapshot {
	return entity.Snapshot{ID: id}
}
