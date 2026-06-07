// Eval для settings.exclude GID-111.
package service

import (
	"context"

	"excluded/domain/model"
)

type Snapshot struct{}

// Исключён как "Snapshot.SnapshotPtr" — удобно сразу отдать указатель.
func (s *Snapshot) SnapshotPtr(ctx context.Context, id string) (*model.Snapshot, error) {
	return nil, nil
}

// Не исключён — репортится.
func (s *Snapshot) Other(ctx context.Context, id string) (*model.Snapshot, error) { // want `GID-111: output data must be returned by value\. Fix: use model\.Snapshot`
	return nil, nil
}
