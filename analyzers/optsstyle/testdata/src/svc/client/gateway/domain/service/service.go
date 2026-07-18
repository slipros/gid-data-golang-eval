// Non-applicability (GID-152): a nested "domain/service" below another layer
// (a client-side package at client/gateway/domain/service) is NOT the
// domain/service layer — the "within" scope is anchored to the module root
// (first segment "client"). An Options parameter by value here is not flagged.
package service

type Options struct {
	Retries int
}

func NewBad(opts Options) int { // not in scope — no diagnostic
	return opts.Retries
}
