// Eval: service конвертирует model <-> entity, но от репозитория
// зависит через интерфейс, а не импортом реализации.
package service

import (
	"svc/dal/entity"
	"svc/dal/repository" // want `GID-132: пакету "svc/domain/service" запрещён импорт "svc/dal/repository" — сервис зависит от репозитория через интерфейс рядом с потребителем`

	"svc/domain/model"
)

// Негатив (граница): импорт entity сервису разрешён — конвертация.
type Snapshot struct {
	repo *repository.Snapshot
}

func (s *Snapshot) Snapshot(id string) (model.Snapshot, error) {
	out, err := s.repo.Snapshot(id)
	if err != nil {
		return model.Snapshot{}, err
	}
	return fromEntity(&out), nil
}

func fromEntity(in *entity.Snapshot) model.Snapshot {
	return model.Snapshot{ID: in.ID}
}
