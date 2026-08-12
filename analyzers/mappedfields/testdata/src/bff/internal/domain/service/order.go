// Eval for GID-266: a gRPC client call in a BFF carries a MappedFields option.
package service

import (
	"context"

	gdgrpcerror "genproto/error"
	"genproto/orderpb"
	gdmapper "helper/mapper"

	"google.golang.org/grpc"
)

var orderFields = gdmapper.MappedFields{
	gdmapper.NewMappedFieldStringEqualWithWholePart("segment_id", "segmentId"),
}

// Catalog — a client interface that does not forward the call options: there is
// nowhere to pass the mapping, so the rule has nothing to ask for.
type Catalog interface {
	Items(ctx context.Context) ([]string, error)
}

type Order struct {
	client  orderpb.OrderServiceClient
	catalog Catalog
}

// --- Class 1: positive ---

func (o *Order) CreateBare(ctx context.Context, in *orderpb.CreateOrderRequest) (*orderpb.Order, error) {
	return o.client.CreateOrder(ctx, in) // want "GID-266: the gRPC call CreateOrder carries no MappedFields option"
}

func (o *Order) OrderWithOtherOption(ctx context.Context, id string) (*orderpb.Order, error) {
	req := orderpb.GetOrderRequest{ID: id}

	return o.client.Order(ctx, &req, grpc.WaitForReady(true)) // want "GID-266: the gRPC call Order carries no MappedFields option"
}

func (o *Order) CreateNilMapping(ctx context.Context, in *orderpb.CreateOrderRequest) (*orderpb.Order, error) {
	return o.client.CreateOrder(ctx, in, gdgrpcerror.WithMappedFields(nil)) // want "GID-266: the MappedFields option of the gRPC call CreateOrder is empty"
}

func (o *Order) CreateEmptyMapping(ctx context.Context, in *orderpb.CreateOrderRequest) (*orderpb.Order, error) {
	return o.client.CreateOrder(ctx, in, gdgrpcerror.WithMappedFields(gdmapper.MappedFields{})) // want "GID-266: the MappedFields option of the gRPC call CreateOrder is empty"
}

func (o *Order) CreateEmptyOptionLiteral(ctx context.Context, in *orderpb.CreateOrderRequest) (*orderpb.Order, error) {
	return o.client.CreateOrder(ctx, in, gdgrpcerror.MappedFieldsInterceptorCallOption{}) // want "GID-266: the MappedFields option of the gRPC call CreateOrder is empty"
}

// CreateFromLiteral — the request is built in the call and carries a field, so
// the callee can reject that field by its own name.
func (o *Order) CreateFromLiteral(ctx context.Context, name string) (*orderpb.Order, error) {
	return o.client.CreateOrder(ctx, &orderpb.CreateOrderRequest{Name: name}) // want "GID-266: the gRPC call CreateOrder carries no MappedFields option"
}

// --- Class 2: negative ---

func (o *Order) CreateMapped(ctx context.Context, in *orderpb.CreateOrderRequest) (*orderpb.Order, error) {
	return o.client.CreateOrder(ctx, in, gdgrpcerror.WithMappedFields(orderFields))
}

func (o *Order) CreateInlineMapping(ctx context.Context, in *orderpb.CreateOrderRequest) (*orderpb.Order, error) {
	return o.client.CreateOrder(ctx, in, gdgrpcerror.WithMappedFields(gdmapper.MappedFields{
		gdmapper.NewMappedFieldStringEqualWithWholePart("name", "title"),
	}))
}

// CreateOptionInVariable — the option is recognised by the name of its type, so
// a value prepared beforehand counts as well as a fresh call.
func (o *Order) CreateOptionInVariable(ctx context.Context, in *orderpb.CreateOrderRequest) (*orderpb.Order, error) {
	option := gdgrpcerror.WithMappedFields(orderFields)

	return o.client.CreateOrder(ctx, in, option)
}

// --- Class 3: boundary ---

// CreateAfterOtherOption — the mapping does not have to come first.
func (o *Order) CreateAfterOtherOption(ctx context.Context, in *orderpb.CreateOrderRequest) (*orderpb.Order, error) {
	return o.client.CreateOrder(ctx, in, grpc.WaitForReady(true), gdgrpcerror.WithMappedFields(orderFields))
}

// CreateSpreadOptions — the caller spreads a prepared slice: the option may be
// inside it, and the call is left alone.
func (o *Order) CreateSpreadOptions(
	ctx context.Context,
	in *orderpb.CreateOrderRequest,
	opts []grpc.CallOption,
) (*orderpb.Order, error) {
	return o.client.CreateOrder(ctx, in, opts...)
}

// Items — a client interface without call options cannot be judged.
func (o *Order) Items(ctx context.Context) ([]string, error) {
	return o.catalog.Items(ctx)
}

// Ping — an RPC taking nothing beside the context sends no field to be rejected.
func (o *Order) Ping(ctx context.Context) error {
	return o.client.Ping(ctx)
}

// DeleteNilRequest — a nil request holds no field either.
func (o *Order) DeleteNilRequest(ctx context.Context) error {
	return o.client.DeleteOrder(ctx, nil)
}

// CreateEmptyRequest — an empty literal is the same case spelled out.
func (o *Order) CreateEmptyRequest(ctx context.Context) (*orderpb.Order, error) {
	return o.client.CreateOrder(ctx, &orderpb.CreateOrderRequest{})
}

// CreateRequestInVariable — what the variable holds is not followed, so the call
// stays judged. This is the shape most of lk-api is written in.
func (o *Order) CreateRequestInVariable(ctx context.Context, name string) (*orderpb.Order, error) {
	req := orderpb.CreateOrderRequest{Name: name}

	return o.client.CreateOrder(ctx, &req) // want "GID-266: the gRPC call CreateOrder carries no MappedFields option"
}

// Describe — an ordinary variadic call is not a gRPC call.
func (o *Order) Describe(ctx context.Context, in *orderpb.CreateOrderRequest) string {
	return join(in.Name, "order")
}

func join(parts ...string) string {
	var out string
	for _, part := range parts {
		out += part
	}

	return out
}
