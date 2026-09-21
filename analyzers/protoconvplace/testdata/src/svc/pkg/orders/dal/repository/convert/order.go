package convert

import (
	"svc/genproto/orderpb"
	"svc/pkg/orders/dal/entity"
)

// Negative: protobuf/entity conversion is allowed in dal/repository/convert.
func OrderFromGRPC(order *orderpb.Order) entity.Order {
	return entity.Order{
		ID:     order.ID,
		Status: order.Status,
		Title:  order.Title,
	}
}
