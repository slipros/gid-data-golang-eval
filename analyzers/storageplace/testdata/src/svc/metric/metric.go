// Boundary: the metrics collector of a driver library is not storage access —
// it is wired in /metric and /app.
package metric

import redisprom "github.com/redis/go-redis/pkg/prometheus"

type Prometheus struct {
	Redis *redisprom.Metrics
}
