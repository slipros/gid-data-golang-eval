// Eval для GID-178 (запрет встраивания sync.Mutex/sync.RWMutex).
package embedmutex

import (
	"sync"

	syncalias "sync"
)

// --- Позитивные кейсы (встраивание ловится) ---

// Встроенный sync.Mutex.
type Cache struct {
	sync.Mutex // want `GID-178: sync\.Mutex is embedded in the struct\. Fix: use a named mutex field \(mu sync\.Mutex\), otherwise Lock/Unlock leak into the type's API`
	data       map[string]string
}

// Встроенный указатель *sync.RWMutex.
type Registry struct {
	*sync.RWMutex // want `GID-178: sync\.RWMutex is embedded in the struct\. Fix: use a named mutex field \(mu sync\.Mutex\), otherwise Lock/Unlock leak into the type's API`
	items         []int
}

// Встраивание через алиас импорта пакета sync — детект по типу, не по тексту.
type Aliased struct {
	syncalias.Mutex // want `GID-178: sync\.Mutex is embedded in the struct\. Fix: use a named mutex field \(mu sync\.Mutex\), otherwise Lock/Unlock leak into the type's API`
	n               int
}

// --- Негативные кейсы (чистый код проходит) ---

// Именованное неэкспортируемое поле — каноническая форма.
type Good struct {
	mu   sync.Mutex
	data map[string]string
}

// Именованное поле-указатель тоже допустимо.
type GoodPtr struct {
	mu *sync.RWMutex
}

// --- Граничные кейсы (похоже, но не матчится) ---

// Свой тип Mutex (не из пакета sync) — встраивание допустимо.
type Mutex struct{}

type WithOwnMutex struct {
	Mutex // не sync.Mutex
}

// sync.WaitGroup — другой тип из sync, не мьютекс.
type WithWaitGroup struct {
	sync.WaitGroup
	done bool
}
