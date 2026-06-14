// Eval for GID-215 (no-inline-entity-literal) in the domain layer.
package service

import (
	"svc/dal/entity"
	"svc/dal/entity/filter"
	"svc/domain/model"
)

// --- Class 1: positive (inline-filling an entity is forbidden) ---

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

// --- Class 2: negative (clean code passes) ---

func emptySnapshot() entity.Snapshot {
	return entity.Snapshot{} // empty literal — zero value, allowed.
}

func modelSnapshot(id, name string) model.Snapshot {
	return model.Snapshot{ID: id, Name: name} // model in domain is normal.
}

func modelCreate(name string) model.CreateSnapshot {
	return model.CreateSnapshot{Name: name}
}

// --- Class 3: boundary ---

// A nested entity literal inside a flagged outer one — a single diagnostic.
func nestedSlice() entity.Snapshots {
	return entity.Snapshots{ // want `GID-215: inline-filling the entity type entity\.Snapshots in the domain layer is forbidden`
		entity.Snapshot{ID: "a"},
		entity.Snapshot{ID: "b"},
	}
}

// map[string]entity.X{...} — the map literal itself is not an entity type (not flagged),
// but the value entity.Snapshot{...} is flagged. An element without an explicit type
// ({ID: id}) also has type entity.Snapshot and is flagged as a non-empty literal.
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
