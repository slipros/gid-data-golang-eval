// Non-applicability: the import is exempted through settings.exclude-packages —
// a default prefix matches it, but it carries no storage access.
package tools

import prom "github.com/redis/go-redis/pkg/instrumentation"

type Tools struct {
	metrics *prom.Metrics
}
