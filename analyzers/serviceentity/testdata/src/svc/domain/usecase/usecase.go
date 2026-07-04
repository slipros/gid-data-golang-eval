// Eval for GID-236 non-applicability: the same shape outside /domain/service
// is not checked — a usecase legitimately orchestrates several entities.
package usecase

import "context"

type JobRepository interface {
	Job(ctx context.Context, id string) (string, error)
}

type Upload struct {
	jobs JobRepository
}
