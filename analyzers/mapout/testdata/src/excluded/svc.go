// Eval of GID-257 settings.exclude: the same violation as in svc/resolver.go,
// cleared only because the method is on the exclusion list
// ("Resolver.legacyResolveChunk").
package excluded

// Cabinet is the value being resolved.
type Cabinet struct {
	ID string
}

// Resolver owns the excluded method.
type Resolver struct{}

func (r *Resolver) legacyResolveChunk(chunk []string, result map[string]Cabinet) {
	for _, id := range chunk {
		result[id] = Cabinet{ID: id}
	}
}
