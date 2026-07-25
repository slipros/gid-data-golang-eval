// Non-applicability: /job added through settings.allow — the driver is fine here.
package sweep

import redis "github.com/redis/go-redis/v9"

type Sweeper struct {
	client *redis.Client
}
