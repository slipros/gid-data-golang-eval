// Eval for GID-261 (no-branch-dispatch).
package service

import "context"

// IntegrationRepository — the dependency the service dispatches over.
type IntegrationRepository interface {
	Integration(ctx context.Context, id string) (string, error)
	IntegrationByOrganization(ctx context.Context, id, org string) (string, error)
	IntegrationsByFilter(ctx context.Context, filter string) ([]string, error)
}

// Integration — the service under test.
type Integration struct {
	repo     IntegrationRepository
	coreRepo IntegrationRepository
}

// --- Positive: assignment form — the incident shape ---

func (i *Integration) Get(ctx context.Context, id, org string) (string, error) {
	var (
		e   string
		err error
	)

	if org == "" { // want `GID-261: method Integration\.Get picks between i\.repo\.Integration and i\.repo\.IntegrationByOrganization by a condition on its input — one method, several operations\. Fix: split it into a method per query \(Integration and IntegrationByOrganization\), each taking the arguments its own query needs, and let the caller choose`
		e, err = i.repo.Integration(ctx, id)
	} else {
		e, err = i.repo.IntegrationByOrganization(ctx, id, org)
	}
	if err != nil {
		return "", err
	}

	return e, nil
}

// --- Positive: return form ---

func (i *Integration) Read(ctx context.Context, id, org string) (string, error) {
	if org == "" { // want `GID-261: method Integration\.Read picks between i\.repo\.Integration and i\.repo\.IntegrationByOrganization`
		return i.repo.Integration(ctx, id)
	} else {
		return i.repo.IntegrationByOrganization(ctx, id, org)
	}
}

// --- Negative: the same method with different arguments is one operation ---

func (i *Integration) SameMethod(ctx context.Context, id, fallback string) (string, error) {
	var (
		e   string
		err error
	)

	if id == "" {
		e, err = i.repo.Integration(ctx, fallback)
	} else {
		e, err = i.repo.Integration(ctx, id)
	}

	return e, err
}

// --- Negative: different receivers — two dependencies, not two ways into one ---

func (i *Integration) TwoDeps(ctx context.Context, id string, core bool) (string, error) {
	var (
		e   string
		err error
	)

	if core {
		e, err = i.coreRepo.Integration(ctx, id)
	} else {
		e, err = i.repo.Integration(ctx, id)
	}

	return e, err
}

// --- Boundary: a branch doing more than the call is not a dispatch arm ---

func (i *Integration) WithSideWork(ctx context.Context, id, org string) (string, error) {
	var (
		e   string
		err error
	)

	if org == "" {
		e, err = i.repo.Integration(ctx, id)
	} else {
		filtered, ferr := i.repo.IntegrationsByFilter(ctx, org)
		if ferr != nil {
			return "", ferr
		}
		e = filtered[0]
	}

	return e, err
}

// --- Positive: an else-if chain is the same defect, only bigger — reported
// once, on the outermost if ---

func (i *Integration) Chain(ctx context.Context, id, org string) (string, error) {
	var (
		e   string
		err error
	)

	if org == "" { // want `GID-261: method Integration\.Chain picks between i\.repo\.Integration, i\.repo\.IntegrationByOrganization and i\.repo\.Integration by a condition on its input`
		e, err = i.repo.Integration(ctx, id)
	} else if id == "" {
		e, err = i.repo.IntegrationByOrganization(ctx, id, org)
	} else {
		e, err = i.repo.Integration(ctx, id)
	}

	return e, err
}
