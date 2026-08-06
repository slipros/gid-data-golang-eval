// Eval of GID-258 settings.exclude: soft degradation as a stated functional
// requirement — the outage must NOT surface as an error, the caller substitutes
// a default. The same violation as in svc/…/service.go, cleared only because
// the method is on the exclusion list ("Resolver.degradeSilently").
package service

import "fmt"

// Registry is the injected dependency that can fail.
type Registry interface {
	ByIDs(ids []string) ([]string, error)
}

// Resolver owns the excluded method.
type Resolver struct {
	registry Registry
}

func (r *Resolver) degradeSilently(ids []string) {
	if _, err := r.registry.ByIDs(ids); err != nil {
		fmt.Println("registry unavailable, degrading to id")
	}
}
