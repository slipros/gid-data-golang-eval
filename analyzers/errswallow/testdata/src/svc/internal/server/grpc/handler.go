// Non-applicability: outside /domain/** the signature is not ours to change —
// a transport handler answers to a foreign contract (an HTTP handler, a Kafka
// message handler, a gRPC interceptor) and has nowhere to return an error to.
// The very shape reported in /domain/service stays clean here.
package grpc

import "fmt"

// Sender is the transport dependency that can fail.
type Sender interface {
	Send(payload string) error
}

// Handler owns the method below.
type Handler struct {
	sender Sender
}

func (h *Handler) Handle(payload string) {
	if err := h.sender.Send(payload); err != nil {
		fmt.Println("send failed")
	}
}
