// Eval для settings.exclude: разрешённое исключение — иногда gRPC
// вызывается прямо в service.
package service

import (
	"google.golang.org/grpc" // want `GID-160: direct import of google\.golang\.org/grpc in the domain layer is forbidden`

	"excluded/pkg/api/orderpb" // исключён через settings.exclude — не флагуется
)

type Order struct {
	conn   *grpc.ClientConn
	client *orderpb.OrderClient
}
