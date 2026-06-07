// Граничный (GID-170): domain импортирует пакеты с похожими, но
// другими сегментами — "events" (мн. число) и "event-api" — диагностики нет.
package boundary

import (
	"svc/event-api/contract"
	"svc/events/dto"
)

type Snapshot struct{}

func (b *Snapshot) FromEvents(in dto.SnapshotDTO) string  { return in.ID }
func (b *Snapshot) FromAPI(in contract.SnapshotContract) string { return in.ID }
