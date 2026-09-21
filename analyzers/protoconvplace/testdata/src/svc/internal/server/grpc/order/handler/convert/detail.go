package convert

import (
	"svc/genproto/orderpb"
	"svc/internal/domain/model"
)

// Negative: substantial conversion is allowed in the leaf convert package.
func CreateOrderFromProto(req *orderpb.CreateOrderRequest) model.CreateOrder {
	return model.CreateOrder{
		Title:  req.Title,
		Status: req.Status,
	}
}
