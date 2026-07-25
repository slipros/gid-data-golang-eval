// Negative: the driver in /dal/repository — exactly where it belongs.
package repository

import redis "github.com/redis/go-redis/v9"

type Repo struct {
	client *redis.Client
}
