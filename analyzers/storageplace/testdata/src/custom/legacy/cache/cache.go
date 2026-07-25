// Non-applicability: the package is exempted through settings.exclude-paths.
package cache

import redis "github.com/redis/go-redis/v9"

type Cache struct {
	client *redis.Client
}
