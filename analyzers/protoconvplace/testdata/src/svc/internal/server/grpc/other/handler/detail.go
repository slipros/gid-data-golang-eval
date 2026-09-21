package handler

import (
	"svc/genproto/orderpb"
	"svc/internal/domain/model"
)

// Non-applicability: only the service handler path is governed by GID-277.
func createOrderFromProto(req *orderpb.CreateOrderRequest) model.CreateOrder {
	return model.CreateOrder{
		Title:  req.Title,
		Status: req.Status,
	}
}
