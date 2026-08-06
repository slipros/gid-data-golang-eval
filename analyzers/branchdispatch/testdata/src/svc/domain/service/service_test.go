// Non-applicability: a _test.go double dispatching over its own state is test
// scaffolding (GID-250).
package service

import "context"

type repoDouble struct {
	byOrg bool
	inner IntegrationRepository
}

func (r *repoDouble) Integration(ctx context.Context, id string) (string, error) {
	var (
		e   string
		err error
	)

	if r.byOrg {
		e, err = r.inner.IntegrationByOrganization(ctx, id, "org")
	} else {
		e, err = r.inner.Integration(ctx, id)
	}

	return e, err
}
