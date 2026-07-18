// Eval for GID-151 (service-model API).
package service

import (
	"context"

	"svc/dal/entity"
	"svc/domain/model"
	nestedentity "svc/server/grpc/dal/entity"
)

type Snapshot struct{}

// --- Positive cases: an entity in an exported method signature ---

func (s *Snapshot) CreateSnapshot(ctx context.Context, in *entity.CreateSnapshot) error { // want `GID-151: method "CreateSnapshot" uses the entity type entity\.CreateSnapshot \(parameter\)\. Fix: the service API takes and returns model, convert to entity internally`
	return nil
}

func (s *Snapshot) SnapshotRaw(ctx context.Context, id string) (entity.Snapshot, error) { // want `GID-151: method "SnapshotRaw" uses the entity type entity\.Snapshot \(result\)`
	return entity.Snapshot{}, nil
}

// Boundary case: the entity is hidden in a named slice type.
func (s *Snapshot) SnapshotsRaw(ctx context.Context) (entity.Snapshots, error) { // want `GID-151: method "SnapshotsRaw" uses the entity type entity\.Snapshots \(result\)`
	return nil, nil
}

// Boundary case: an entity inside a map.
func (s *Snapshot) SnapshotsByID(ctx context.Context) (map[string]*entity.Snapshot, error) { // want `GID-151: method "SnapshotsByID" uses the entity type entity\.Snapshot \(result\)`
	return nil, nil
}

// --- Negative cases: an API on model ---

func (s *Snapshot) Snapshot(ctx context.Context, id string) (model.Snapshot, error) {
	var out entity.Snapshot // an entity inside the body is the norm
	return fromEntity(&out), nil
}

func (s *Snapshot) Create(ctx context.Context, in *model.CreateSnapshot) error {
	return nil
}

// Boundary case: server/grpc/dal/entity is a package that merely contains
// the segments "dal", "entity" nested below another layer (server/grpc) —
// it is not the real /dal/entity layer, so referencing it must NOT be
// flagged (would be a false positive under path Contains).
func (s *Snapshot) SnapshotNested(ctx context.Context, id string) (nestedentity.FakeEntity, error) {
	return nestedentity.FakeEntity{}, nil
}

// --- Not applicable: unexported helpers (conversion) ---

func (s *Snapshot) toEntity(in *model.CreateSnapshot) entity.CreateSnapshot {
	return entity.CreateSnapshot{Name: in.Name}
}

func fromEntity(in *entity.Snapshot) model.Snapshot {
	return model.Snapshot{ID: in.ID}
}
