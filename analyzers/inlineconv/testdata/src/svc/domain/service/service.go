// Eval для GID-215 (no-inline-entity-literal) в domain-слое.
package service

import (
	"svc/dal/entity"
	"svc/dal/entity/filter"
	"svc/domain/model"
)

// --- Класс 1: позитивный (инлайн-заполнение entity запрещено) ---

func createSnapshot(name string) entity.CreateSnapshot {
	return entity.CreateSnapshot{Name: name} // want `GID-215: inline-filling the entity type entity\.CreateSnapshot in the domain layer is forbidden\. Fix: put conversion in a convert package \(<Dst><Type>From<Src>\)`
}

func snapshotPtr(id string) *entity.Snapshot {
	return &entity.Snapshot{ID: id} // want `GID-215: inline-filling the entity type entity\.Snapshot in the domain layer is forbidden`
}

func snapshotSlice() entity.Snapshots {
	return entity.Snapshots{{ID: "a"}, {ID: "b"}} // want `GID-215: inline-filling the entity type entity\.Snapshots in the domain layer is forbidden`
}

func snapshotsFilter(name string) filter.Snapshots {
	return filter.Snapshots{Name: name, Limit: 10} // want `GID-215: inline-filling the entity type filter\.Snapshots in the domain layer is forbidden`
}

// --- Класс 2: негативный (чистый код проходит) ---

func emptySnapshot() entity.Snapshot {
	return entity.Snapshot{} // пустой литерал — zero value, разрешён.
}

func modelSnapshot(id, name string) model.Snapshot {
	return model.Snapshot{ID: id, Name: name} // model в domain — норма.
}

func modelCreate(name string) model.CreateSnapshot {
	return model.CreateSnapshot{Name: name}
}

// --- Класс 3: граничный ---

// Вложенный entity-литерал внутри зафлаганного внешнего — одна диагностика.
func nestedSlice() entity.Snapshots {
	return entity.Snapshots{ // want `GID-215: inline-filling the entity type entity\.Snapshots in the domain layer is forbidden`
		entity.Snapshot{ID: "a"},
		entity.Snapshot{ID: "b"},
	}
}

// map[string]entity.X{...} — сам map-литерал не entity-тип (не флагается),
// а вот значение entity.Snapshot{...} — флагается. Элемент без явного типа
// ({ID: id}) тоже имеет тип entity.Snapshot и флагается как непустой литерал.
func snapshotMap(id string) map[string]entity.Snapshot {
	return map[string]entity.Snapshot{
		id: {ID: id}, // want `GID-215: inline-filling the entity type entity\.Snapshot in the domain layer is forbidden`
	}
}

func snapshotMapExplicit(id string) map[string]entity.Snapshot {
	return map[string]entity.Snapshot{
		id: entity.Snapshot{ID: id}, // want `GID-215: inline-filling the entity type entity\.Snapshot in the domain layer is forbidden`
	}
}
