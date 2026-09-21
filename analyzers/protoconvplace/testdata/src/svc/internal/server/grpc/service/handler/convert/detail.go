package convert

import (
	"svc/genproto/orderpb"
	"svc/internal/domain/model"
)

// Negative: substantial conversion is allowed in the exact handler/convert package.
func CreateOrderFromProto(req *orderpb.CreateOrderRequest) model.CreateOrder {
	return model.CreateOrder{
		Title:  req.Title,
		Status: req.Status,
	}
}

// Negative: the reverse direction is allowed there too.
func OrderToProto(order *model.Order) *orderpb.Order {
	return &orderpb.Order{
		ID:     order.ID,
		Status: order.Status,
		Title:  order.Title,
	}
}
