// Позитив (GID-170): domain не зависит от event-слоя.
package notifier

import (
	"svc/app"               // want `GID-225: package "svc/domain/notifier" must not import "svc/app"\. Fix: the composition root and transport are leaves; nobody imports them`
	"svc/domain/model"
	"svc/event/dto"         // want `GID-170: package "svc/domain/notifier" must not import "svc/event/dto"\. Fix: domain does not depend on the event layer; event converts model <-> DTO, not the other way`
	"svc/server/middleware" // want `GID-225: package "svc/domain/notifier" must not import "svc/server/middleware"\. Fix: the composition root and transport are leaves; nobody imports them`
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
