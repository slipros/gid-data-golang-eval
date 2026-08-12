// Stub of a generated pb client: every RPC ends in opts ...grpc.CallOption.
package orderpb

import (
	"context"

	"google.golang.org/grpc"
)

type CreateOrderRequest struct {
	Name string
}

type GetOrderRequest struct {
	ID string
}

type DeleteOrderRequest struct {
	ID string
}

type Order struct {
	ID string
}

type OrderServiceClient interface {
	CreateOrder(ctx context.Context, in *CreateOrderRequest, opts ...grpc.CallOption) (*Order, error)
	Order(ctx context.Context, in *GetOrderRequest, opts ...grpc.CallOption) (*Order, error)
	DeleteOrder(ctx context.Context, in *DeleteOrderRequest, opts ...grpc.CallOption) error
	Ping(ctx context.Context, opts ...grpc.CallOption) error
}
