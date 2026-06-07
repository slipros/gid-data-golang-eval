// Негатив: кэширующий репозиторий в /dal/repository — здесь кэшу и место.
// Оборачивает основной репозиторий прямой ссылкой, без интерфейса.
package repository

import (
	redis "github.com/redis/go-redis/v9"
)

type Snapshot struct{}

func (s *Snapshot) Snapshot(id string) (string, error) { return id, nil }

// CachedSnapshot — вся магия с кэшом живёт здесь.
type CachedSnapshot struct {
	repo  *Snapshot // прямая ссылка на основной репозиторий
	cache *redis.Client
}

func (c *CachedSnapshot) Snapshot(id string) (string, error) {
	return c.repo.Snapshot(id)
}
