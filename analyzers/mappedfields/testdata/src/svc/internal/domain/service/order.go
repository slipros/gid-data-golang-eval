// Class 4, non-applicability: the module owns a data layer (internal/dal), so
// it is not a BFF — it reaches another service through a repository (GID-160),
// and the mapping of a foreign validation error is not its shape: no diagnostic.
package service

import (
	"context"

	"genproto/orderpb"
)

type Order struct {
	client orderpb.OrderServiceClient
}

func (o *Order) Create(ctx context.Context, name string) (*orderpb.Order, error) {
	return o.client.CreateOrder(ctx, &orderpb.CreateOrderRequest{Name: name})
}
