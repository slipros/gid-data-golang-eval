// Eval of GID-246: a struct whose name carries "adapter" is needless
// indirection. Default settings — no exclusions.
package wiring

import "context"

// --- Class 1: positive — struct names carrying "adapter" ---

// DedupAdapter — CamelCase "Adapter" suffix.
type DedupAdapter struct { // want `GID-246: "DedupAdapter" is an adapter struct`
	hits int
}

func (d *DedupAdapter) GetResult(ctx context.Context, key string) (string, bool, error) {
	_ = ctx
	d.hits++
	return key, true, nil
}

// adapterCache — lowercase, "adapter" as a substring, not a suffix.
type adapterCache struct { // want `GID-246: "adapterCache" is an adapter struct`
	n int
}

// --- Class 2: negative ---

// MetricsObserver carries behavior but no "adapter" in its name.
type MetricsObserver struct {
	count int
}

func (m *MetricsObserver) Observe(name string) { _ = name; m.count++ }

// deps is a plain wiring holder.
type deps struct {
	dedup *DedupAdapter
	cache *adapterCache
}

// AdapterFunc is a func type, not a struct — the rule only scopes struct types.
type AdapterFunc func(ctx context.Context) error

// New is the wiring function.
func New() *deps {
	return &deps{dedup: &DedupAdapter{}, cache: &adapterCache{}}
}

// --- Class 3: boundary — an interface named Adapter is a consumer-side port ---

// CacheAdapter is an interface (a port declared on the consumer side), not a
// struct — legitimate, not flagged.
type CacheAdapter interface {
	GetResult(ctx context.Context, key string) (string, bool, error)
}
