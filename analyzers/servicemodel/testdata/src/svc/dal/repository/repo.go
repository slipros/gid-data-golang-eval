// Not applicable: the rule applies only in /domain/service —
// for a repository entity in signatures is mandatory.
package repository

import (
	"context"

	"svc/dal/entity"
)

type Snapshot struct{}

func (s *Snapshot) Snapshot(ctx context.Context, id string) (entity.Snapshot, error) {
	return entity.Snapshot{ID: id}, nil
}
