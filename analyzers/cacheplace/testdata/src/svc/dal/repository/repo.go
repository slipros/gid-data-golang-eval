// Negative: a caching repository in /dal/repository — this is where the cache belongs.
// It wraps the main repository by a direct reference, without an interface.
package repository

import (
	redis "github.com/redis/go-redis/v9"
)

type Snapshot struct{}

func (s *Snapshot) Snapshot(id string) (string, error) { return id, nil }

// CachedSnapshot — all the cache magic lives here.
type CachedSnapshot struct {
	repo  *Snapshot // a direct reference to the main repository
	cache *redis.Client
}

func (c *CachedSnapshot) Snapshot(id string) (string, error) {
	return c.repo.Snapshot(id)
}
