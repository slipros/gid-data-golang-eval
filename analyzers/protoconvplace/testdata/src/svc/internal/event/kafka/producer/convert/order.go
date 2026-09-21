package convert

import (
	"svc/genproto/orderpb"
	"svc/internal/domain/model"
)

// Negative: event conversion is allowed in the exact producer/convert package.
func OrderEventFromModel(order model.Order) *orderpb.Order {
	return &orderpb.Order{
		ID:     order.ID,
		Status: order.Status,
		Title:  order.Title,
	}
}
