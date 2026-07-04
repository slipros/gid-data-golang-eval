// Package repository — a dal/repository interface used to eval the
// "declared in another package" negative case (GID-134 scope, GID-236).
package repository

import "context"

// FileRepository — declared outside /domain/service; the same-package check
// of GID-236 does not see it.
type FileRepository interface {
	File(ctx context.Context, id string) (string, error)
}
