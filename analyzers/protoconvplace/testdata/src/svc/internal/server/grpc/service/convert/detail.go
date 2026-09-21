package convert

import (
	"svc/genproto/orderpb"
	"svc/internal/domain/model"
)

// Positive: a sibling convert package is not the handler-owned destination.
func createOrderFromProto( // want `GID-277: function "createOrderFromProto" performs substantial cross-representation conversion outside /server/grpc/service/handler/convert\. Fix: move the field mapping to /server/grpc/service/handler/convert and call it from "convert"`
	req *orderpb.CreateOrderRequest,
) model.CreateOrder {
	return model.CreateOrder{
		Title:  req.Title,
		Status: req.Status,
	}
}
