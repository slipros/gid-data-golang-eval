// Eval for GID-111 (service).
package service

import (
	"context"

	"svc/domain/model"
)

type Snapshot struct{}

// --- Positive: input by value ---

func (s *Snapshot) Create(ctx context.Context, in model.CreateSnapshot) error { // want `GID-111: input data must be passed by pointer\. Fix: use \*model\.CreateSnapshot`
	return nil
}

// --- Positive: output by pointer ---

func (s *Snapshot) Snapshot(ctx context.Context, id string) (*model.Snapshot, error) { // want `GID-111: output data must be returned by value\. Fix: use model\.Snapshot`
	return nil, nil
}

// --- Negative: canonical — input *T, output T ---

func (s *Snapshot) Update(ctx context.Context, in *model.CreateSnapshot) error {
	return nil
}

func (s *Snapshot) Get(ctx context.Context, id string) (model.Snapshot, error) {
	return model.Snapshot{}, nil
}

// Edge case: a named string type — not a struct, by value is fine.
func (s *Snapshot) Status(ctx context.Context, st model.SnapshotStatus) error {
	return nil
}

// Edge case: a slice of structs — a slice header, by value is fine.
func (s *Snapshot) Many(ctx context.Context, in []model.CreateSnapshot) error {
	return nil
}

// Not applicable: unexported helpers are not checked.
func (s *Snapshot) helper(in model.CreateSnapshot) model.Snapshot {
	return model.Snapshot{}
}
