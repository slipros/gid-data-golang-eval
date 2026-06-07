// Класс 2 (негативный) — convert-пакет сервиса: конвертация здесь и живёт,
// инлайн-литералы entity разрешены.
package convert

import (
	"svc/dal/entity"
	"svc/domain/model"
)

func EntityCreateSnapshotFromModel(in model.CreateSnapshot) entity.CreateSnapshot {
	return entity.CreateSnapshot{Name: in.Name}
}

func ModelSnapshotFromEntity(in entity.Snapshot) model.Snapshot {
	return model.Snapshot{ID: in.ID, Name: in.Name}
}

func EntitySnapshotsFromModel() entity.Snapshots {
	return entity.Snapshots{entity.Snapshot{ID: "a"}}
}
