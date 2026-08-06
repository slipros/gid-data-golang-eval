// Eval of the GID-261 settings: settings.exclude clears a method by name
// ("Legacy") or by Type.Method ("Integration.Get").
package service

import "context"

type IntegrationRepository interface {
	Integration(ctx context.Context, id string) (string, error)
	IntegrationByOrganization(ctx context.Context, id, org string) (string, error)
}

type Integration struct{ repo IntegrationRepository }

// Excluded as "Integration.Get".
func (i *Integration) Get(ctx context.Context, id, org string) (string, error) {
	if org == "" {
		return i.repo.Integration(ctx, id)
	} else {
		return i.repo.IntegrationByOrganization(ctx, id, org)
	}
}

// Excluded as "Legacy" — a bare method name clears it on any type.
func (i *Integration) Legacy(ctx context.Context, id, org string) (string, error) {
	if org == "" {
		return i.repo.Integration(ctx, id)
	} else {
		return i.repo.IntegrationByOrganization(ctx, id, org)
	}
}

// Not excluded — still reported.
func (i *Integration) Read(ctx context.Context, id, org string) (string, error) {
	if org == "" { // want `GID-261: method Integration\.Read picks between i\.repo\.Integration and i\.repo\.IntegrationByOrganization`
		return i.repo.Integration(ctx, id)
	} else {
		return i.repo.IntegrationByOrganization(ctx, id, org)
	}
}
