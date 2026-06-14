// Eval for settings.exclude: an allowed exception — sometimes gRPC
// is called directly in a service.
package service

import (
	"google.golang.org/grpc" // want `GID-160: direct import of google\.golang\.org/grpc in the domain layer is forbidden`

	"excluded/pkg/api/orderpb" // excluded via settings.exclude — not flagged
)

type Order struct {
	conn   *grpc.ClientConn
	client *orderpb.OrderClient
}
