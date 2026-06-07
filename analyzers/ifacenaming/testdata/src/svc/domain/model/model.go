// Eval GID-173 — неприменимость: /domain/model вне scope.
package model

import "context"

// Голая роль вне scope — диагностики быть не должно.
type Repository interface {
	Hello(ctx context.Context) error
}
