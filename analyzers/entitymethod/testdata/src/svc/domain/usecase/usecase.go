// Not applicable: /domain/usecase is out of scope (the scope is repository and service).
package usecase

import "context"

type Job struct{}

// A List prefix, a ByID suffix, a method without an entity — all legal in usecase: out of scope.
func (u *Job) ListJobs(ctx context.Context) error {
	return nil
}

func (u *Job) Fetch(ctx context.Context) error {
	return nil
}
