// Eval для GID-111 (service).
package service

import (
	"context"

	"svc/domain/model"
)

type Snapshot struct{}

// --- Позитив: вход по значению ---

func (s *Snapshot) Create(ctx context.Context, in model.CreateSnapshot) error { // want `GID-111: input data must be passed by pointer\. Fix: use \*model\.CreateSnapshot`
	return nil
}

// --- Позитив: выход по указателю ---

func (s *Snapshot) Snapshot(ctx context.Context, id string) (*model.Snapshot, error) { // want `GID-111: output data must be returned by value\. Fix: use model\.Snapshot`
	return nil, nil
}

// --- Негатив: канон — вход *T, выход T ---

func (s *Snapshot) Update(ctx context.Context, in *model.CreateSnapshot) error {
	return nil
}

func (s *Snapshot) Get(ctx context.Context, id string) (model.Snapshot, error) {
	return model.Snapshot{}, nil
}

// Граничный кейс: именованный строковый тип — не структура, по значению норма.
func (s *Snapshot) Status(ctx context.Context, st model.SnapshotStatus) error {
	return nil
}

// Граничный кейс: слайс структур — заголовок слайса, по значению норма.
func (s *Snapshot) Many(ctx context.Context, in []model.CreateSnapshot) error {
	return nil
}

// Неприменимость: неэкспортируемые хелперы не проверяются.
func (s *Snapshot) helper(in model.CreateSnapshot) model.Snapshot {
	return model.Snapshot{}
}
