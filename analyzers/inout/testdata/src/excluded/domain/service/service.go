// Eval for settings.exclude GID-111.
package service

import (
	"context"

	"excluded/domain/model"
)

type Snapshot struct{}

// Excluded as "Snapshot.SnapshotPtr" — convenient to return a pointer directly.
func (s *Snapshot) SnapshotPtr(ctx context.Context, id string) (*model.Snapshot, error) {
	return nil, nil
}

// Not excluded — reported.
func (s *Snapshot) Other(ctx context.Context, id string) (*model.Snapshot, error) { // want `GID-111: output data must be returned by value\. Fix: use model\.Snapshot`
	return nil, nil
}
