// Positive (default settings): /job is not an allowed layer.
package sweep

import redis "github.com/redis/go-redis/v9" // want `GID-249: package "svc/job/sweep" reaches a data store directly — driver "github.com/redis/go-redis/v9" belongs to the repository layer\. Fix: move the storage code to /dal/repository`

type Sweeper struct {
	client *redis.Client
}
