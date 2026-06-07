// Позитив (GID-170): domain не зависит от event-слоя.
package notifier

import (
	"svc/app"               // want `GID-225: пакету "svc/domain/notifier" запрещён импорт "svc/app" — composition root и транспорт — листья: их никто не импортирует`
	"svc/domain/model"
	"svc/event/dto"         // want `GID-170: пакету "svc/domain/notifier" запрещён импорт "svc/event/dto" — domain не зависит от event-слоя: event конвертирует model <-> DTO, не наоборот`
	"svc/server/middleware" // want `GID-225: пакету "svc/domain/notifier" запрещён импорт "svc/server/middleware" — composition root и транспорт — листья: их никто не импортирует`
)

type Snapshot struct{}

// Негатив: model в domain — норма.
func (n *Snapshot) Build() model.Snapshot {
	return model.Snapshot{}
}

// Позитив выше: event-DTO в domain-слое запрещён.
func (n *Snapshot) leak(in dto.SnapshotDTO) {}

// Позитивы выше (GID-225): composition root и транспорт — листья.
func (n *Snapshot) leakLeaves() {
	app.Wire()
	middleware.Noop()
}
