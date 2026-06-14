// Not applicable: the package has no structs — the rule has nothing to check.
package nostructs

import "sync"

// A function with an embedded mutex? No — the mutex lives in a named variable.
func New() *sync.Mutex {
	var mu sync.Mutex
	return &mu
}

// An interface — mutex embedding does not happen in interfaces, and there is none here.
type Locker interface {
	Lock()
	Unlock()
}
