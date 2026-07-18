// Eval of GID-159 (boundary): a package with a "domain" segment nested under
// a DIFFERENT layer (server/grpc) — a business-domain grouping, not the
// module's domain layer. Before the pathseg.HasLayer fix,
// pathseg.Contains(pkgPath, "domain") matched "domain" anywhere in the path
// and falsely banned this import as a domain-layer cache import. With
// HasLayer (anchored to the module root, which here is "server", not
// "domain") the rule does not apply — no diagnostic.
package domain

import (
	redis "github.com/redis/go-redis/v9"
)

// Handler — importing the cache library here must NOT be flagged: the
// package is not in the domain layer (it is nested under server/grpc).
type Handler struct {
	cache *redis.Client
}
