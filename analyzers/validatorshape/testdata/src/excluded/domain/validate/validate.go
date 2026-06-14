// Eval GID-213: a type from settings.exclude is not considered a validator.
package validate

import "context"

type jobReq struct{ ID string }

// Listed in settings.exclude as "HealthCheck" — not flagged without Validate.
type HealthCheck struct{}

// A correct validator — not flagged in any case.
type Ping struct{}

func (v *Ping) Validate(ctx context.Context, req *jobReq) error { return nil }

// Not in exclude and without a correct Validate — flagged; verifies that the
// exclusion does not silence the whole package.
type Echo struct{} // want `GID-213: validator "Echo" must have a Validate\(ctx context.Context, req \*T\) error method\. Fix: add it`
