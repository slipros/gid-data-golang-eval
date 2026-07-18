// Eval of GID-246 positive: the rule fires repository-wide, not only in the app
// layer — a struct named Adapter here is flagged too (default settings). Exempt
// legitimate infrastructure adapters via settings.exclude-paths.
package dedup

import "context"

// Adapter realizes a port over an external cache.
type Adapter struct { // want `GID-246: "Adapter" is an adapter struct`
	hits int
}

func (a *Adapter) GetResult(ctx context.Context, key string) (string, bool, error) {
	_ = ctx
	a.hits++
	return key, true, nil
}
