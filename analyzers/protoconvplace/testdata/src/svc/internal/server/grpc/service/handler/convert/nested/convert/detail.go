package convert

import (
	"svc/genproto/orderpb"
	"svc/internal/domain/model"
)

// Positive: a descendant named convert is not the exact handler/convert package.
func createOrderFromProto( // want `GID-277: function "createOrderFromProto" performs substantial cross-representation conversion outside /server/grpc/service/handler/convert\. Fix: move the field mapping to /server/grpc/service/handler/convert and call it from "convert"`
	req *orderpb.CreateOrderRequest,
) model.CreateOrder {
	return model.CreateOrder{
		Title:  req.Title,
		Status: req.Status,
	}
}
