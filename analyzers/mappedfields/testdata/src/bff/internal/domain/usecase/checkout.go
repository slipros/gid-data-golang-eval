// The usecase layer is judged the same way the service layer is.
package usecase

import (
	"context"

	gdgrpcerror "genproto/error"
	"genproto/orderpb"
	gdmapper "helper/mapper"
)

type Checkout struct {
	client orderpb.OrderServiceClient
}

func (c *Checkout) Create(ctx context.Context, in *orderpb.CreateOrderRequest) (*orderpb.Order, error) {
	return c.client.CreateOrder(ctx, in) // want "GID-266: the gRPC call CreateOrder carries no MappedFields option"
}

func (c *Checkout) CreateMapped(ctx context.Context, in *orderpb.CreateOrderRequest) (*orderpb.Order, error) {
	return c.client.CreateOrder(ctx, in, gdgrpcerror.WithMappedFields(gdmapper.MappedFields{
		gdmapper.NewMappedFieldStringEqualWithWholePart("name", "title"),
	}))
}
