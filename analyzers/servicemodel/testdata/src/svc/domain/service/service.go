// Eval для GID-151 (service-model API).
package service

import (
	"context"

	"svc/dal/entity"
	"svc/domain/model"
)

type Snapshot struct{}

// --- Позитивные кейсы: entity в сигнатуре экспортируемого метода ---

func (s *Snapshot) CreateSnapshot(ctx context.Context, in *entity.CreateSnapshot) error { // want `GID-151: method "CreateSnapshot" uses the entity type entity\.CreateSnapshot \(parameter\)\. Fix: the service API takes and returns model, convert to entity internally`
	return nil
}

func (s *Snapshot) SnapshotRaw(ctx context.Context, id string) (entity.Snapshot, error) { // want `GID-151: method "SnapshotRaw" uses the entity type entity\.Snapshot \(result\)`
	return entity.Snapshot{}, nil
}

// Граничный кейс: entity спрятана в слайсе именованного типа.
func (s *Snapshot) SnapshotsRaw(ctx context.Context) (entity.Snapshots, error) { // want `GID-151: method "SnapshotsRaw" uses the entity type entity\.Snapshots \(result\)`
	return nil, nil
}

// Граничный кейс: entity внутри мапы.
func (s *Snapshot) SnapshotsByID(ctx context.Context) (map[string]*entity.Snapshot, error) { // want `GID-151: method "SnapshotsByID" uses the entity type entity\.Snapshot \(result\)`
	return nil, nil
}

// --- Негативные кейсы: API на model ---

func (s *Snapshot) Snapshot(ctx context.Context, id string) (model.Snapshot, error) {
	var out entity.Snapshot // entity внутри тела — норма
	return fromEntity(&out), nil
}

func (s *Snapshot) Create(ctx context.Context, in *model.CreateSnapshot) error {
	return nil
}

// --- Неприменимость: неэкспортируемые хелперы (конвертация) ---

func (s *Snapshot) toEntity(in *model.CreateSnapshot) entity.CreateSnapshot {
	return entity.CreateSnapshot{Name: in.Name}
}

func fromEntity(in *entity.Snapshot) model.Snapshot {
	return model.Snapshot{ID: in.ID}
}
