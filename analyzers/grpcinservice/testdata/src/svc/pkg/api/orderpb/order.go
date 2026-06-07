// Stub сгенерированного pb-пакета: импортирует grpc — значит gRPC-backed.
package orderpb

import "google.golang.org/grpc"

type Order struct{ ID string }

type OrderClient struct{ cc *grpc.ClientConn }

func NewOrderClient(cc *grpc.ClientConn) *OrderClient { return &OrderClient{cc: cc} }
