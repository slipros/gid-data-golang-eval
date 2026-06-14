// Eval of GID-173 — /server/** in scope: the bare role Service.
package grpc

import "context"

// --- Positive case: a bare role in /server/** ---

type Service interface { // want `GID-173: interface "Service" must be named with an entity prefix\. Fix: e\.g\. HelloRepository`
	Hello(ctx context.Context) error
}

// --- Negative case ---

type HelloService interface {
	Hello(ctx context.Context) error
}
