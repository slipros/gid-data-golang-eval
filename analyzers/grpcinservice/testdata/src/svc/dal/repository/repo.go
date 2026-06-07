// Негатив (неприменимость): repository — правильное место для gRPC-вызовов.
package repository

import (
	"google.golang.org/grpc"

	"svc/pkg/api/orderpb"
)

type Order struct {
	client *orderpb.OrderClient
}

func NewOrder(cc *grpc.ClientConn) *Order {
	return &Order{client: orderpb.NewOrderClient(cc)}
}
