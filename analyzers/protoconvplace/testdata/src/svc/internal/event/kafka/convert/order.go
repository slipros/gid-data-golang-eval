package convert

import (
	"svc/genproto/orderpb"
	"svc/internal/domain/model"
)

// Positive: a sibling convert package is not the producer-owned destination.
func orderEventFromModel( // want `GID-277: function "orderEventFromModel" performs substantial cross-representation conversion outside /event/kafka/producer/convert\. Fix: move the field mapping to /event/kafka/producer/convert and call it from "convert"`
	order model.Order,
) *orderpb.Order {
	return &orderpb.Order{
		ID:     order.ID,
		Status: order.Status,
		Title:  order.Title,
	}
}
