// Positive: a key-value driver in /client — this package is in fact a
// key-value repository and belongs in /dal/repository.
package ratelimit

import redis "github.com/redis/go-redis/v9" // want `GID-249: package "svc/client/ratelimit" reaches a data store directly — driver "github.com/redis/go-redis/v9" belongs to the repository layer\. Fix: move the storage code to /dal/repository`

type Limiter struct {
	client *redis.Client
}
