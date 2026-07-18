// Boundary: a package literally named "app" nested under domain/service is NOT
// the composition root — pathseg.HasLayer anchors the "app" exception to the
// module root, unlike pathseg.Contains which would match "app" anywhere in the
// path. Without the fix, a bare Options struct here would be silently exempted
// from GID-126; with the fix the rule correctly applies.
package app

type Options struct { // want `GID-126: an options type must have an entity prefix\. Fix: use JobOptions, not bare Options`
	Retries int
}
