// Positive (GID-275): the package name is read off the path the way goimports
// reads it — a /vN suffix skipped, a go- prefix trimmed, cut at the first
// character that cannot be in an identifier.
package app

import (
	billing "example.com/svc/internal/client/billing/v2" // want `GID-275: import alias billing is not needed — nothing else in the file is called billing\. Fix: import it as "example\.com/svc/internal/client/billing/v2" and refer to it as billing`
	ads "example.com/svc/internal/client/go-ads"         // want `GID-275: import alias ads is not needed — nothing else in the file is called ads\. Fix: import it as "example\.com/svc/internal/client/go-ads" and refer to it as ads`
	stats "example.com/svc/internal/client/stats.v3"     // want `GID-275: import alias stats is not needed — nothing else in the file is called stats\. Fix: import it as "example\.com/svc/internal/client/stats\.v3" and refer to it as stats`
)

var _ = billing.X + ads.X + stats.X
