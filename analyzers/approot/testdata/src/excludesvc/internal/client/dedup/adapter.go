// Eval of GID-246 settings.exclude-paths: the "client" segment is excluded, so
// legitimate infrastructure adapters here are not flagged.
package dedup

import "context"

// Adapter is a real infrastructure adapter — exempt because its package sits
// under an excluded path ("client").
type Adapter struct {
	hits int
}

func (a *Adapter) GetResult(ctx context.Context, key string) (string, bool, error) {
	_ = ctx
	a.hits++
	return key, true, nil
}
