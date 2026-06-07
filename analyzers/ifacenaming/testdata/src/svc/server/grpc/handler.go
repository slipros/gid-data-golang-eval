// Eval GID-173 — /server/** в scope: голая роль Service.
package grpc

import "context"

// --- Позитивный кейс: голая роль в /server/** ---

type Service interface { // want `GID-173: интерфейс "Service" именуется с префиксом сущности \(например, HelloRepository\)`
	Hello(ctx context.Context) error
}

// --- Негативный кейс ---

type HelloService interface {
	Hello(ctx context.Context) error
}
