// Class 4, non-applicability: the transport layer is out of scope — the rule
// judges /domain/service and /domain/usecase, mirroring GID-160.
package http

import (
	"context"

	"genproto/orderpb"
)

type Handler struct {
	client orderpb.OrderServiceClient
}

func (h *Handler) Create(ctx context.Context, name string) (*orderpb.Order, error) {
	return h.client.CreateOrder(ctx, &orderpb.CreateOrderRequest{Name: name})
}
