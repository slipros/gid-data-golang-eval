package service

import "context"

type Invoice struct {
	repo InvoiceRepository
}

func NewInvoice(repo InvoiceRepository) *Invoice {
	return &Invoice{repo: repo}
}

func (i *Invoice) Invoice(ctx context.Context) error {
	return i.repo.Invoice(ctx)
}

type InvoiceRepository interface { // want `GID-276: interface InvoiceRepository is declared below type Invoice \(line 5\)`
	Invoice(ctx context.Context) error
}
