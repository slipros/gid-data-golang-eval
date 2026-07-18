// Eval boundary GID-210: a domain/model segment nested under another layer
// (server/grpc) is NOT the model layer — the layer is anchored to the module
// root (pathseg.HasLayer), so a Create struct here with generated fields must
// NOT be flagged, unlike domainsvc/domain/model itself.
package model

import "time"

// Would be flagged if the layer segment were matched anywhere in the path
// (pathseg.Contains) instead of being anchored to the module root.
type CreateJob struct {
	ID        int
	CreatedAt time.Time
	UpdatedAt time.Time
}
