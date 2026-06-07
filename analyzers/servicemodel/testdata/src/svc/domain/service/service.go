// Eval для GID-151 (service-model API).
package service

import (
	"context"

	"svc/dal/entity"
	"svc/domain/model"
)

type Snapshot struct{}

// --- Позитивные кейсы: entity в сигнатуре экспортируемого метода ---

func (s *Snapshot) CreateSnapshot(ctx context.Context, in *entity.CreateSnapshot) error { // want `GID-151: метод "CreateSnapshot" использует entity-тип entity\.CreateSnapshot \(параметр\) — API сервиса принимает и возвращает model, конвертация в entity выполняется внутри`
	return nil
}

func (s *Snapshot) SnapshotRaw(ctx context.Context, id string) (entity.Snapshot, error) { // want `GID-151: метод "SnapshotRaw" использует entity-тип entity\.Snapshot \(результат\)`
	return entity.Snapshot{}, nil
}

// Граничный кейс: entity спрятана в слайсе именованного типа.
func (s *Snapshot) SnapshotsRaw(ctx context.Context) (entity.Snapshots, error) { // want `GID-151: метод "SnapshotsRaw" использует entity-тип entity\.Snapshots \(результат\)`
	return nil, nil
}

// Граничный кейс: entity внутри мапы.
func (s *Snapshot) SnapshotsByID(ctx context.Context) (map[string]*entity.Snapshot, error) { // want `GID-151: метод "SnapshotsByID" использует entity-тип entity\.Snapshot \(результат\)`
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
