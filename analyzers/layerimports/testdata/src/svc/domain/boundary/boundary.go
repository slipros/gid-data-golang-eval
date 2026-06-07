// Граничный (GID-170/226): domain импортирует пакеты с похожими, но
// другими сегментами — "events", "event-api", "metrics" — диагностики нет.
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
