// Eval GID-230: types from settings.exclude are not treated as handlers.
package handler

import "context"

type jobReq struct{ ID string }
type jobResp struct{ OK bool }

// Listed in settings.exclude as "HealthCheck" — no Handle, not flagged.
type HealthCheck struct{}

// Canonical handler — never flagged.
type Run struct{}

func (h *Run) Handle(ctx context.Context, req *jobReq) (*jobResp, error) { return nil, nil }

// Not in exclude and without Handle — flagged; proves the exclusion does not
// silence the whole package.
type Echo struct{} // want `GID-230: handler "Echo" must have a Handle method with context\.Context as the first param and error as the last result\. Fix: func \(h \*Echo\) Handle\(ctx context\.Context, req \*rpc\.Request\) \(\*rpc\.Response, error\)`
