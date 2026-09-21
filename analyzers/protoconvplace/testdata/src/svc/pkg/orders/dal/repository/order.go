package repository

import (
	"svc/genproto/orderpb"
	"svc/pkg/orders/dal/entity"
)

// Positive: protobuf/entity mapping belongs in dal/repository/convert.
func orderFromGRPC( // want `GID-277: function "orderFromGRPC" performs substantial cross-representation conversion outside /dal/repository/convert\. Fix: move the field mapping to /dal/repository/convert and call it from "repository"`
	order *orderpb.Order,
) entity.Order {
	return entity.Order{
		ID:     order.ID,
		Status: order.Status,
		Title:  order.Title,
	}
}
