// Eval of GID-171 (boundary): a package with a "domain" segment nested under
// a DIFFERENT layer (server/grpc) — a business-domain grouping, not the
// module's domain layer. Before the pathseg.HasLayer fix,
// pathseg.Contains(pkgPath, "domain") matched "domain" anywhere in the path
// and falsely classified this server-layer package as the domain layer. With
// HasLayer (anchored to the module root, which here is "server", not
// "domain") the rule does not apply — no diagnostic.
package domain

// FilterOrders — a filter struct that must NOT be flagged here: the package
// is not in the domain layer (it is nested under server/grpc).
type FilterOrders struct {
	Status string
}
