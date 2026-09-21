package other

import (
	"svc/genproto/orderpb"
	"svc/internal/domain/model"
)

// Non-applicability: packages outside transport/event are not judged.
func createOrderFromProto(req *orderpb.CreateOrderRequest) model.CreateOrder {
	return model.CreateOrder{
		Title:  req.Title,
		Status: req.Status,
	}
}
