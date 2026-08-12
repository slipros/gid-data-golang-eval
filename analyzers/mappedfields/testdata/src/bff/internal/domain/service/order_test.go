// Class 4, non-applicability: a _test.go file is not judged. The double
// repeats the signatures of the client interface it fakes (GID-250 keeps it in
// this package), and the test calls it with the arguments its assertion needs —
// not with the options production code owes the frontend.
package service

import (
	"context"
	"testing"

	"genproto/orderpb"

	"google.golang.org/grpc"
)

type orderClientStub struct {
	order orderpb.Order
}

func (s *orderClientStub) CreateOrder(
	_ context.Context,
	_ *orderpb.CreateOrderRequest,
	_ ...grpc.CallOption,
) (*orderpb.Order, error) {
	return &s.order, nil
}

func (s *orderClientStub) Order(
	_ context.Context,
	_ *orderpb.GetOrderRequest,
	_ ...grpc.CallOption,
) (*orderpb.Order, error) {
	return &s.order, nil
}

func (s *orderClientStub) DeleteOrder(
	_ context.Context,
	_ *orderpb.DeleteOrderRequest,
	_ ...grpc.CallOption,
) error {
	return nil
}

func (s *orderClientStub) Ping(_ context.Context, _ ...grpc.CallOption) error { return nil }

func TestOrderCreateBare(t *testing.T) {
	client := orderClientStub{order: orderpb.Order{ID: "1"}}
	svc := Order{client: &client}

	got, err := svc.client.CreateOrder(context.Background(), &orderpb.CreateOrderRequest{Name: "x"})
	if err != nil || got.ID != "1" {
		t.Fatal("unexpected result")
	}
}
