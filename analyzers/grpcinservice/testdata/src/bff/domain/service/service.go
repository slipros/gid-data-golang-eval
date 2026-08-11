// Eval for GID-160, non-applicability: a BFF module has no data layer, so
// there is no repository to move the gRPC call into and the rule stays silent.
package service

import (
	"google.golang.org/grpc"

	"svc/pkg/api/orderpb"
)

type Order struct {
	conn   *grpc.ClientConn
	client *orderpb.OrderClient
}
