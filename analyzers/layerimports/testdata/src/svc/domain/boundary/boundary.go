// Boundary (GID-170/226): domain imports packages with similar but
// different segments — "events", "event-api", "metrics" — there is no diagnostic.
package boundary

import (
	"svc/event-api/contract"
	"svc/events/dto"
	"svc/metrics/registry"
)

type Snapshot struct{}

func (b *Snapshot) FromEvents(in dto.SnapshotDTO) string  { return in.ID }
func (b *Snapshot) FromAPI(in contract.SnapshotContract) string { return in.ID }
func (b *Snapshot) FromRegistry(in registry.Registry) *registry.Registry { return &in }
