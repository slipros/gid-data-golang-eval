// Eval for GID-215 (boundary): inline-filling a type from a package with
// segments dal and entity nested under a DIFFERENT layer (server/api) must
// NOT be flagged. Before the pathseg.HasLayer fix,
// pathseg.Contains(pkg.Path(), "dal", "entity") matched "dal"+"entity"
// anywhere in the foreign package's path and falsely treated
// svc/server/api/dal/entity as the module's entity layer. With HasLayer
// (anchored to the module root, which here is "server", not "dal") the type
// is not an entity-layer type — no diagnostic.
package service

import (
	miscentity "svc/server/api/dal/entity"
)

func miscValue(v string) miscentity.Misc {
	return miscentity.Misc{Value: v}
}
