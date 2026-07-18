// Boundary: a package literally named "app" nested under domain/service is NOT
// the composition root (internal/app) — pathseg.HasLayer anchors the "app"
// exception to the module root, unlike pathseg.Contains which would match
// "app" anywhere in the path. Without the fix this bare New() would be
// silently exempted from GID-104; with the fix the rule correctly applies.
package app

type Worker struct{}

func New() *Worker { // want `GID-104: a constructor must be named New<Entity>, not bare New\. Fix: rename it to New<Entity> \(bare New clashes with other entities in the package\)`
	return &Worker{}
}
