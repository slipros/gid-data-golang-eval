// Неприменимость: правило действует только в /domain/service —
// у репозитория entity в сигнатурах обязательна.
package repository

import (
	"context"

	"svc/dal/entity"
)

type Snapshot struct{}

func (s *Snapshot) Snapshot(ctx context.Context, id string) (entity.Snapshot, error) {
	return entity.Snapshot{ID: id}, nil
}
