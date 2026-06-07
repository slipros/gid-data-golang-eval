// Eval для GID-159: сервис про кэш не знает.
package service

import (
	redis "github.com/redis/go-redis/v9" // want `GID-159: импорт кэш-библиотеки "github.com/redis/go-redis/v9" в domain-слое запрещён — кэш оформляется кэширующим репозиторием в /dal/repository, оборачивающим основной`
)

type Snapshot struct {
	cache *redis.Client
}
