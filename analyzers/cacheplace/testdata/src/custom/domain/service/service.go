// Eval для settings.packages: кастомная inhouse кэш-библиотека.
package service

import (
	cache "example.com/inhouse/cache" // want `GID-159: importing the cache library "example.com/inhouse/cache" in the domain layer is forbidden`

	redis "github.com/redis/go-redis/v9" // дефолтный список заменён — не флагуется
)

type Snapshot struct {
	hot  *cache.Store
	warm *redis.Client
}
