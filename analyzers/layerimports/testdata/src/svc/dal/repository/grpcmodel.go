// Eval for GID-132 (the ARD-review case): a repository that wraps a gRPC
// client is legal (GID-160 even requires calling gRPC through a repository),
// but its methods must still return entity — exposing internal/domain/model
// from the dal layer is a violation regardless of the data source.
package repository

import (
	"svc/domain/model" // want `GID-132: package "svc/dal/repository" must not import "svc/domain/model"\. Fix: the dal layer works only with entity, domain types are not available to it`
)

// SegmentGRPCClient — the injected gRPC client of an external service.
type SegmentGRPCClient interface {
	Fetch(id string) (string, error)
}

// Segment wraps a gRPC client — that part is canonical.
type Segment struct {
	client SegmentGRPCClient
}

// Segment returns a DOMAIN MODEL from dal — the violation: a repository
// returns entity; entity -> model conversion belongs to the service layer.
func (s *Segment) Segment(id string) (model.Snapshot, error) {
	if _, err := s.client.Fetch(id); err != nil {
		return model.Snapshot{}, err
	}
	return model.Snapshot{}, nil
}
