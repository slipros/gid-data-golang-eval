package usecase

import "context"

type PlaceOrder struct {
	orders PlaceOrderOrderService
}

type PlaceOrderOrderService interface { // want `GID-276: interface PlaceOrderOrderService is declared below type PlaceOrder \(line 5\)`
	Order(ctx context.Context) error
}
