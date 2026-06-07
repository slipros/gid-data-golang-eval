// Eval для settings.exclude: разрешённое исключение — иногда gRPC
// вызывается прямо в service.
package service

import (
	"google.golang.org/grpc" // want `GID-160: прямой импорт google\.golang\.org/grpc в domain-слое запрещён`

	"excluded/pkg/api/orderpb" // исключён через settings.exclude — не флагуется
)

type Order struct {
	conn   *grpc.ClientConn
	client *orderpb.OrderClient
}
