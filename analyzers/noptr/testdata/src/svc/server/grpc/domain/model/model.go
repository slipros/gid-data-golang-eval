// Eval boundary GID-121: a domain/model segment nested under another layer
// (server/grpc) is NOT the model layer — the layer is anchored to the module
// root (pathseg.HasLayer), so a pointer to a simple type here must NOT be
// flagged, unlike svc/domain/model itself. GID-120 (*uuid.UUID) is unaffected
// by scope and is not exercised here.
package model

import "time"

// Would be flagged if the layer segment were matched anywhere in the path
// (pathseg.Contains) instead of being anchored to the module root.
type Snapshot struct {
	CompletedAt *time.Time
}
