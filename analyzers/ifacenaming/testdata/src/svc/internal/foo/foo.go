// Eval of GID-173 — inapplicability: /internal/foo is out of scope.
package foo

import "context"

// A bare role out of scope — there must be no diagnostic.
type Repository interface {
	Hello(ctx context.Context) error
}
