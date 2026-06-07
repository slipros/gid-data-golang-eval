// Eval: service конвертирует model <-> entity, но от репозитория
// зависит через интерфейс, а не импортом реализации.
package service

import (
	"svc/client/billing" // want `GID-228: package "svc/domain/service" must not import "svc/client/billing"\. Fix: service/usecase depend on the client through an interface in domain/model, see GID-134`
	"svc/dal/entity"
	"svc/dal/repository" // want `GID-132: package "svc/domain/service" must not import "svc/dal/repository"\. Fix: a service depends on the repository through an interface next to the consumer`
	"svc/metric"         // want `GID-226: package "svc/domain/service" must not import "svc/metric"\. Fix: domain receives metrics through an interface; the metric package is wired in app`

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

// Позитивы выше: клиент — через интерфейс (GID-228), метрики — интерфейсом (GID-226).
func (s *Snapshot) leakDeps(c *billing.Client, m *metric.Prometheus) {}
