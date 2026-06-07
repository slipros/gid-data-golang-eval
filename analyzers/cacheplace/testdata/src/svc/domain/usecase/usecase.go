// Позитив (граница): in-memory LRU в usecase — тоже кэш, тоже запрещён.
package usecase

import (
	lru "github.com/hashicorp/golang-lru/v2" // want `GID-159: импорт кэш-библиотеки "github.com/hashicorp/golang-lru/v2" в domain-слое запрещён`
)

type Upload struct {
	hot *lru.Cache[string, string]
}
