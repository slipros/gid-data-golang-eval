package handler

import (
	"svc/genproto/orderpb"
	"svc/internal/domain/model"
	handlerconvert "svc/internal/server/grpc/service/handler/convert"
)

// Positive: a substantial protobuf-to-model helper in a handler is conversion.
func createOrderFromProto( // want `GID-277: function "createOrderFromProto" performs substantial cross-representation conversion outside a convert package\. Fix: move the field mapping to a leaf convert package and call it from "handler"`
	req *orderpb.CreateOrderRequest,
) model.CreateOrder {
	// Calling a converter for a subfield or unrelated value must not exempt
	// substantial mapping that still happens locally.
	_ = handlerconvert.CreateOrderFromProto(req)

	return model.CreateOrder{
		Title:  req.Title,
		Status: req.Status,
	}
}

// Negative: complete delegation has no substantial local mapping to report.
func delegatedCreateOrder(req *orderpb.CreateOrderRequest) model.CreateOrder {
	return handlerconvert.CreateOrderFromProto(req)
}

// Boundary: one-field transport outcome wrapping remains handler policy.
func responseFromModel(order *model.Order) *orderpb.Response {
	return &orderpb.Response{Order: &orderpb.Order{ID: order.ID}}
}

// Negative: same-representation construction is not conversion.
func cloneOrder(order *model.Order) model.Order {
	return model.Order{
		ID:     order.ID,
		Status: order.Status,
		Title:  order.Title,
	}
}
