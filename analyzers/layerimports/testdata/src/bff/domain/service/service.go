// Eval for GID-267, non-applicability: a BFF module owns no data layer, so
// there is no repository to move the client call into — calling other services
// and shaping the answer for the frontend IS the business logic here. The
// service reaches the client directly and the rule stays silent (the same gate
// GID-160 uses, modlayout.HasDataLayer). No diagnostic is expected here.
package service

import (
	"bff/client/billing"

	"bff/domain/model"
)

type Invoice struct {
	client *billing.Client
}

// Invoice converts the client's own model into the BFF vocabulary — the
// service API still takes and returns model (GID-151).
func (i *Invoice) Invoice(id string) (model.Invoice, error) {
	out, err := i.client.Invoice(id)
	if err != nil {
		return model.Invoice{}, err
	}

	return model.Invoice{ID: out.ID}, nil
}
