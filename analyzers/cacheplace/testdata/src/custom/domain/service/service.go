// Eval for settings.packages: a custom in-house cache library.
package service

import (
	cache "example.com/inhouse/cache" // want `GID-159: importing the cache library "example.com/inhouse/cache" in the domain layer is forbidden`

	redis "github.com/redis/go-redis/v9" // the default list is replaced — not flagged
)

type Snapshot struct {
	hot  *cache.Store
	warm *redis.Client
}
