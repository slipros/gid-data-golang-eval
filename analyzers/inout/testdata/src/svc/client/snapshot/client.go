// Eval for GID-111 client scope: a client has no domain/model or dal/entity
// of its own, so the rule applies to its own same-module named structs
// instead (client.md — method signatures).
package snapshot

import (
	"context"

	"external/pb"
)

// CreateSnapshotRequest and Snapshot are the client's own request/response types.
type CreateSnapshotRequest struct{ Name string }

type Snapshot struct{ ID string }

type SnapshotStatus string

type Client struct{}

// --- Positive: input by value ---

func (c *Client) Create(ctx context.Context, in CreateSnapshotRequest) error { // want `GID-111: input data must be passed by pointer\. Fix: use \*snapshot\.CreateSnapshotRequest`
	return nil
}

// --- Positive: output by pointer ---

func (c *Client) Get(ctx context.Context, id string) (*Snapshot, error) { // want `GID-111: output data must be returned by value\. Fix: use snapshot\.Snapshot`
	return nil, nil
}

// --- Negative: canonical — input *T, output T ---

func (c *Client) Update(ctx context.Context, in *CreateSnapshotRequest) error {
	return nil
}

func (c *Client) List(ctx context.Context) (Snapshot, error) {
	return Snapshot{}, nil
}

// Edge case: a named string type — not a struct, by value is fine.
func (c *Client) Status(ctx context.Context, st SnapshotStatus) error {
	return nil
}

// Edge case: a slice of structs — a slice header, by value is fine.
func (c *Client) Many(ctx context.Context, in []CreateSnapshotRequest) error {
	return nil
}

// Not applicable: pb.Response is a foreign (non-module) type — a generated/
// third-party dependency, not the client's own model.
func (c *Client) Raw(ctx context.Context) (*pb.Response, error) {
	return nil, nil
}

// Not applicable: unexported methods are not checked.
func (c *Client) helper(in CreateSnapshotRequest) Snapshot {
	return Snapshot{}
}
