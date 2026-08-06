package service

import "fmt"

// --- Non-applicability: a _test.go file is not judged. A test reports its
// failures through *testing.T, not through a result (GID-250), so a helper
// that logs and moves on has no signature to fix. ---

func checkResolve(r *Resolver, ids []string) {
	if _, err := r.registry.ByIDs(ids); err != nil {
		fmt.Println("fixture setup failed")
	}
}
