// Class 4, non-applicability: a library module — no composition root, no data
// layer — is not a BFF: it answers to no frontend, so there is no request
// vocabulary to map a foreign validation error into: no diagnostic.
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
