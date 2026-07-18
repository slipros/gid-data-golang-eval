// Boundary case: domain/service nested under another layer (server/grpc) is
// NOT the domain/service layer — pathseg.HasLayer anchors the layer to the
// module root, so an anonymous tx-signature field here must NOT be flagged
// (would be a false positive under a plain path Contains).
package nested

import "context"

type Service struct {
	tx func(ctx context.Context, fn func(ctx context.Context) error) error
}
