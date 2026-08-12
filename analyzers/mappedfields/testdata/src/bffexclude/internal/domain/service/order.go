// Class 4, non-applicability: settings.exclude names the RPC being called, so
// DeleteOrder and OrderServiceClient.Order need no mapping — the sibling
// CreateOrder on the exclusion-free list is still reported, which is what makes
// the exclusion visible.
package service

import (
	"context"

	"genproto/orderpb"
)

type Order struct {
	client orderpb.OrderServiceClient
}

func (o *Order) Delete(ctx context.Context, id string) error {
	return o.client.DeleteOrder(ctx, &orderpb.DeleteOrderRequest{ID: id})
}

func (o *Order) Order(ctx context.Context, id string) (*orderpb.Order, error) {
	return o.client.Order(ctx, &orderpb.GetOrderRequest{ID: id})
}

func (o *Order) Create(ctx context.Context, name string) (*orderpb.Order, error) {
	return o.client.CreateOrder(ctx, &orderpb.CreateOrderRequest{Name: name}) // want "GID-266: the gRPC call CreateOrder carries no MappedFields option"
}
