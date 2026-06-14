// Positive (GID-170): domain does not depend on the event layer.
package notifier

import (
	"svc/app"               // want `GID-225: package "svc/domain/notifier" must not import "svc/app"\. Fix: the composition root and transport are leaves; nobody imports them`
	"svc/domain/model"
	"svc/event/dto"         // want `GID-170: package "svc/domain/notifier" must not import "svc/event/dto"\. Fix: domain does not depend on the event layer; event converts model <-> DTO, not the other way`
	"svc/server/middleware" // want `GID-225: package "svc/domain/notifier" must not import "svc/server/middleware"\. Fix: the composition root and transport are leaves; nobody imports them`
)

type Snapshot struct{}

// Negative: model in domain is fine.
func (n *Snapshot) Build() model.Snapshot {
	return model.Snapshot{}
}

// Positive above: event DTO is forbidden in the domain layer.
func (n *Snapshot) leak(in dto.SnapshotDTO) {}

// Positives above (GID-225): the composition root and transport are leaves.
func (n *Snapshot) leakLeaves() {
	app.Wire()
	middleware.Noop()
}
