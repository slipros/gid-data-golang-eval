// Eval of GID-246 settings.exclude-paths: the "legacy" segment is listed as the
// escape hatch, so an adapter under it is not flagged — for a concrete proven
// case, not because any layer is privileged (an adapter in /internal/client IS
// flagged, see approotsvc).
package legacy

import "context"

// Adapter is legacy that can't be rewritten right now — exempt because its
// package sits under an excluded path ("legacy").
type Adapter struct {
	hits int
}

func (a *Adapter) GetResult(ctx context.Context, key string) (string, bool, error) {
	_ = ctx
	a.hits++
	return key, true, nil
}
