// Non-applicability: a usecase is not judged — the rule scope is /domain/service.
package usecase

import "context"

type IntegrationRepository interface {
	Integration(ctx context.Context, id string) (string, error)
	IntegrationByOrganization(ctx context.Context, id, org string) (string, error)
}

type Integration struct{ repo IntegrationRepository }

func (i *Integration) Get(ctx context.Context, id, org string) (string, error) {
	var (
		e   string
		err error
	)

	if org == "" {
		e, err = i.repo.Integration(ctx, id)
	} else {
		e, err = i.repo.IntegrationByOrganization(ctx, id, org)
	}

	return e, err
}
