// Eval для GID-159: сервис про кэш не знает.
package service

import (
	redis "github.com/redis/go-redis/v9" // want `GID-159: importing the cache library "github.com/redis/go-redis/v9" in the domain layer is forbidden\. Fix: implement caching as a caching repository in /dal/repository that wraps the main one`
)

type Snapshot struct {
	cache *redis.Client
}
