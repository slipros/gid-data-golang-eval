// Package grpc — the service's "own" package (the /server/grpc layer), NOT
// model and NOT the same package as the repository/service consumer.
// Interfaces from here must not be used in foreign packages — define them
// next to the consumer.
package grpc

// Notifier — an interface of a foreign service package (the server layer).
type Notifier interface {
	Notify(msg string) error
}
