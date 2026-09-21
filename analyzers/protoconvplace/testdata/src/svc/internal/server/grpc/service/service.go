package service

import (
	"svc/genproto/orderpb"
	"svc/internal/domain/model"
)

// Positive: conversion in service wiring is outside handler/convert.
func createOrderFromProto( // want `GID-277: function "createOrderFromProto" performs substantial cross-representation conversion outside /server/grpc/service/handler/convert\. Fix: move the field mapping to /server/grpc/service/handler/convert and call it from "service"`
	req *orderpb.CreateOrderRequest,
) model.CreateOrder {
	return model.CreateOrder{
		Title:  req.Title,
		Status: req.Status,
	}
}
