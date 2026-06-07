// Eval для settings.packages: кастомная inhouse кэш-библиотека.
package service

import (
	cache "example.com/inhouse/cache" // want `GID-159: импорт кэш-библиотеки "example.com/inhouse/cache" в domain-слое запрещён`

	redis "github.com/redis/go-redis/v9" // дефолтный список заменён — не флагуется
)

type Snapshot struct {
	hot  *cache.Store
	warm *redis.Client
}
