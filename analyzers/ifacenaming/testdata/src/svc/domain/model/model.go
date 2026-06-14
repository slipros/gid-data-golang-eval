// Eval of GID-173 — inapplicability: /domain/model is out of scope.
package model

import "context"

// A bare role out of scope — there must be no diagnostic.
type Repository interface {
	Hello(ctx context.Context) error
}
