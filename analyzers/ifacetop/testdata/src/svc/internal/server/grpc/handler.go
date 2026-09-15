package grpc

type Handler struct {
	svc HandlerOrderService
}

type HandlerOrderService interface {
	Order() error
}
