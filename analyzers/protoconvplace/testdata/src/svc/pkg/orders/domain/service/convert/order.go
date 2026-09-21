package convert

import (
	"svc/genproto/orderpb"
	"svc/pkg/orders/domain/model"
)

// Negative: outbound gRPC-client conversion is allowed in domain/service/convert.
func OrderFromGRPC(order *orderpb.Order) model.Order {
	return model.Order{
		ID:     order.ID,
		Status: order.Status,
		Title:  order.Title,
	}
}
