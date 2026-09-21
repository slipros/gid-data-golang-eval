package producer

import (
	"svc/genproto/orderpb"
	"svc/internal/domain/model"
)

// Positive: event adapters also delegate substantial payload conversion.
func orderEventFromModel( // want `GID-277: function "orderEventFromModel" performs substantial cross-representation conversion outside a convert package`
	order model.Order,
) *orderpb.Order {
	return &orderpb.Order{
		ID:     order.ID,
		Status: order.Status,
		Title:  order.Title,
	}
}
