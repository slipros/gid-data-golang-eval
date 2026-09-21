package service

import (
	"svc/genproto/orderpb"
	"svc/pkg/orders/domain/model"
)

// Positive: outbound gRPC-client mapping belongs in domain/service/convert.
func orderFromGRPC( // want `GID-277: function "orderFromGRPC" performs substantial cross-representation conversion outside /domain/service/convert\. Fix: move the field mapping to /domain/service/convert and call it from "service"`
	order *orderpb.Order,
) model.Order {
	return model.Order{
		ID:     order.ID,
		Status: order.Status,
		Title:  order.Title,
	}
}
