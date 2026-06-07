// Eval: repository работает только с entity.
package repository

import (
	"svc/dal/entity"

	"svc/domain/model" // want `GID-132: пакету "svc/dal/repository" запрещён импорт "svc/domain/model" — dal-слой работает только с entity, domain-типы ему недоступны`
)

type Snapshot struct{}

// Негатив: entity в repo — норма.
func (s *Snapshot) Snapshot(id string) (entity.Snapshot, error) {
	return entity.Snapshot{ID: id}, nil
}

// Позитив выше: model в dal-слое запрещён.
func (s *Snapshot) leak(in *model.Snapshot) {}
