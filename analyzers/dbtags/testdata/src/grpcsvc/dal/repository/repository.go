// Eval for GID-125: the repository of this module talks gRPC, not SQL.
package repository

import (
	"google.golang.org/grpc"

	"grpcsvc/dal/entity"
)

// Documents fetches documents over gRPC.
type Documents struct {
	conn *grpc.ClientConn
}

// ByID returns one document.
func (d *Documents) ByID(id string) entity.Document {
	return entity.Document{ID: id}
}
