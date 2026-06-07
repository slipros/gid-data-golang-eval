// Позитив (граница): in-memory LRU в usecase — тоже кэш, тоже запрещён.
package usecase

import (
	lru "github.com/hashicorp/golang-lru/v2" // want `GID-159: importing the cache library "github.com/hashicorp/golang-lru/v2" in the domain layer is forbidden`
)

type Upload struct {
	hot *lru.Cache[string, string]
}
