// Eval boundary GID-234: a domain/model segment nested under another layer
// (server/grpc) is NOT the model layer — the layer is anchored to the module
// root (pathseg.HasLayer), so a generic error name here must NOT be flagged,
// unlike svc/domain/model itself.
package model

import "errors"

// Would be flagged if the layer segment were matched anywhere in the path
// (pathseg.Contains) instead of being anchored to the module root.
var ErrNotFound = errors.New("not found")
