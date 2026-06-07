// Eval для GID-160: service вызывает gRPC через repository.
package service

import (
	"google.golang.org/grpc" // want `GID-160: прямой импорт google\.golang\.org/grpc в domain-слое запрещён — gRPC вызывается через repository \(исключения: nolint или settings\.exclude\)`

	"svc/pkg/api/orderpb" // want `GID-160: импорт gRPC-пакета "svc/pkg/api/orderpb" в domain-слое запрещён — gRPC вызывается через repository`
)

type Order struct {
	conn   *grpc.ClientConn
	client *orderpb.OrderClient
}
