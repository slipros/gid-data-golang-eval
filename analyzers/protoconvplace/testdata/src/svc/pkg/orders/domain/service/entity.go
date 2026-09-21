package service

import (
	"svc/pkg/orders/dal/entity"
	"svc/pkg/orders/domain/model"
)

// Negative: model/entity conversion has no protobuf representation and belongs to GID-215.
func modelFromEntity(order entity.Order) model.Order {
	return model.Order{
		ID:     order.ID,
		Status: order.Status,
		Title:  order.Title,
	}
}
