package consumer

import (
	"svc/genproto/orderpb"
	"svc/internal/domain/model"
)

// Positive: consumer mapping belongs in the consumer-owned convert package.
func orderFromEvent( // want `GID-277: function "orderFromEvent" performs substantial cross-representation conversion outside /event/kafka/consumer/convert\. Fix: move the field mapping to /event/kafka/consumer/convert and call it from "consumer"`
	event *orderpb.Order,
) model.Order {
	return model.Order{
		ID:     event.ID,
		Status: event.Status,
		Title:  event.Title,
	}
}
