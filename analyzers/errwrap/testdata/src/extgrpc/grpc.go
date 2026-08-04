// Package extgrpc stands in for a generated gRPC client's option package: it
// lives outside the module under analysis, so a variadic parameter of its
// CallOption type marks a call into an external system.
package extgrpc

// CallOption — a per-call option of the external client.
type CallOption interface {
	name() string
}
