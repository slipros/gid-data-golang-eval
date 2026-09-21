package handler

import (
	"svc/genproto/orderpb"
	"svc/internal/domain/model"
)

// Non-applicability: test fixture converters are skipped under GID-250.
func fixtureCreateOrderFromProto(req *orderpb.CreateOrderRequest) model.CreateOrder {
	return model.CreateOrder{
		Title:  req.Title,
		Status: req.Status,
	}
}
