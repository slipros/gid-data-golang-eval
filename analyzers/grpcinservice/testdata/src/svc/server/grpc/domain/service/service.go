// Eval boundary GID-160: a domain/service segment nested under another layer
// (server/grpc) is NOT the domain layer itself — the layer is anchored to the
// module root (pathseg.HasLayer), so importing grpc directly here must NOT be
// flagged, unlike svc/domain/service itself.
package service

import "google.golang.org/grpc"

// Would be flagged if the layer segment were matched anywhere in the path
// (pathseg.Contains) instead of being anchored to the module root.
type Order struct {
	conn *grpc.ClientConn
}
