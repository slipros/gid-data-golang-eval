// Eval for GID-159: the service knows nothing about the cache.
package service

import (
	redis "github.com/redis/go-redis/v9" // want `GID-159: importing the cache library "github.com/redis/go-redis/v9" in the domain layer is forbidden\. Fix: implement caching as a caching repository in /dal/repository that wraps the main one`
)

type Snapshot struct {
	cache *redis.Client
}
