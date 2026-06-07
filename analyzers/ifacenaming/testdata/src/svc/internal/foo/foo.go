// Eval GID-173 — неприменимость: /internal/foo вне scope.
package foo

import "context"

// Голая роль вне scope — диагностики быть не должно.
type Repository interface {
	Hello(ctx context.Context) error
}
