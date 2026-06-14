// Eval GID-135: converters live in convert/.
package service

import "context"

type Snapshot struct{ Name string }

type Row struct{ Name string }

// Positive: a converter function in the service package itself.
func ModelSnapshotFromRow(in *Row) Snapshot { // want `GID-135: converter "ModelSnapshotFromRow" must live in a convert/ subpackage of its layer`
	return Snapshot{Name: in.Name}
}

// Edge case: a ctx helper is not a converter (GID-166).
func SessionFromContext(ctx context.Context) (string, bool) {
	s, ok := ctx.Value(struct{}{}).(string)
	return s, ok
}

// Negative: ordinary functions do not match the pattern.
func NewSnapshot() *Snapshot { return &Snapshot{} }
