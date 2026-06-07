// Позитив (GID-224): server импортирует dal и domain/service — запрещено;
// негатив: domain/model и validate транспорту разрешены.
package handler

import (
	"svc/dal/repository" // want `GID-224: пакету "svc/server/http/handler" запрещён импорт "svc/dal/repository" — транспорт работает только с domain/model: сервисы и зависимости инжектятся интерфейсами у потребителя`
	"svc/domain/model"
	"svc/domain/service" // want `GID-224: пакету "svc/server/http/handler" запрещён импорт "svc/domain/service" — транспорт работает только с domain/model: сервисы и зависимости инжектятся интерфейсами у потребителя`
	"svc/validate"
)

type Snapshot struct {
	svc  *service.Snapshot
	repo *repository.Snapshot
	v    *validate.Snapshot
}

// Негатив: model в handler — норма (вход/выход транспорта).
func (h *Snapshot) Get() model.Snapshot {
	return model.Snapshot{}
}
