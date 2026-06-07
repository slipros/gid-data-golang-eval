// Позитив (GID-224): schedule дёргает service напрямую — запрещено;
// негатив: model разрешён.
package sync

import (
	"svc/domain/model"
	"svc/domain/service" // want `GID-224: пакету "svc/schedule/sync" запрещён импорт "svc/domain/service" — транспорт работает только с domain/model: сервисы и зависимости инжектятся интерфейсами у потребителя`
)

type Job struct {
	svc *service.Snapshot
}

// Негатив: model в schedule — норма.
func (j *Job) Run() model.Snapshot {
	return model.Snapshot{}
}
