// Eval for GID-178 (a ban on embedding sync.Mutex/sync.RWMutex).
package embedmutex

import (
	"sync"

	syncalias "sync"
)

// --- Positive cases (embedding is caught) ---

// An embedded sync.Mutex.
type Cache struct {
	sync.Mutex // want `GID-178: sync\.Mutex is embedded in the struct\. Fix: use a named mutex field \(mu sync\.Mutex\), otherwise Lock/Unlock leak into the type's API`
	data       map[string]string
}

// An embedded pointer *sync.RWMutex.
type Registry struct {
	*sync.RWMutex // want `GID-178: sync\.RWMutex is embedded in the struct\. Fix: use a named mutex field \(mu sync\.Mutex\), otherwise Lock/Unlock leak into the type's API`
	items         []int
}

// Embedding via an aliased import of the sync package — detected by type, not by text.
type Aliased struct {
	syncalias.Mutex // want `GID-178: sync\.Mutex is embedded in the struct\. Fix: use a named mutex field \(mu sync\.Mutex\), otherwise Lock/Unlock leak into the type's API`
	n               int
}

// --- Negative cases (clean code passes) ---

// A named unexported field — the canonical form.
type Good struct {
	mu   sync.Mutex
	data map[string]string
}

// A named pointer field is also allowed.
type GoodPtr struct {
	mu *sync.RWMutex
}

// --- Edge cases (similar, but not matched) ---

// A custom Mutex type (not from the sync package) — embedding is allowed.
type Mutex struct{}

type WithOwnMutex struct {
	Mutex // not sync.Mutex
}

// sync.WaitGroup — another type from sync, not a mutex.
type WithWaitGroup struct {
	sync.WaitGroup
	done bool
}
