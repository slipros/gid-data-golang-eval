// Неприменимость: /domain/usecase вне scope (scope — repository и service).
package usecase

import "context"

type Job struct{}

// Префикс List, суффикс ByID, метод без сущности — всё легально в usecase: вне scope.
func (u *Job) ListJobs(ctx context.Context) error {
	return nil
}

func (u *Job) Fetch(ctx context.Context) error {
	return nil
}
