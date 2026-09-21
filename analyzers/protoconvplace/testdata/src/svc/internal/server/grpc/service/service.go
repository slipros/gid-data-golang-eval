package order

import (
	"svc/genproto/orderpb"
	"svc/internal/domain/model"
)

// Non-applicability: gRPC service wiring is outside the handler package.
func createOrderFromProto(req *orderpb.CreateOrderRequest) model.CreateOrder {
	return model.CreateOrder{
		Title:  req.Title,
		Status: req.Status,
	}
}
