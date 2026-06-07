// Класс 4 (неприменимость) — /dal/repository не входит в domain-слой,
// правило не применяется: entity-литералы здесь не флагаются.
package repository

import "svc/dal/entity"

func build(id string) entity.Snapshot {
	return entity.Snapshot{ID: id}
}

func buildSlice() entity.Snapshots {
	return entity.Snapshots{{ID: "a"}}
}
