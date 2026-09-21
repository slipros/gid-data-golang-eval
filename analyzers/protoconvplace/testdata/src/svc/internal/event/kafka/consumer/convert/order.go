package convert

import (
	"svc/genproto/orderpb"
	"svc/internal/domain/model"
)

// Negative: consumer conversion is allowed in the exact consumer/convert package.
func OrderFromEvent(event *orderpb.Order) model.Order {
	return model.Order{
		ID:     event.ID,
		Status: event.Status,
		Title:  event.Title,
	}
}
