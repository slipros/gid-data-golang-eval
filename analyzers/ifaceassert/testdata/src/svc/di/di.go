// Package di — wiring through a container: the constructor is handed to the
// container as a function value, so nothing here converts the type and the
// assertion is what pulls the mismatch back to compile time.
package di

import "svc/usecase"

// The container resolves the dependency by reflection at start-up; the
// assertion is the only compile-time check of the contract.
var _ usecase.LatestPageStore = objectStore{}

type objectStore struct{}

func (objectStore) Object(key string) []byte { return []byte(key) }

// Container — a minimal dependency container.
type Container struct {
	providers []any
}

// Provide registers a constructor.
func (c *Container) Provide(ctor any) { c.providers = append(c.providers, ctor) }

// Wire registers the store constructor.
func Wire(c *Container) {
	c.Provide(newObjectStore)
}

func newObjectStore() objectStore { return objectStore{} }
