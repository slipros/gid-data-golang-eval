// Eval для GID-178 (запрет встраивания sync.Mutex/sync.RWMutex).
package embedmutex

import (
	"sync"

	syncalias "sync"
)

// --- Позитивные кейсы (встраивание ловится) ---

// Встроенный sync.Mutex.
type Cache struct {
	sync.Mutex // want `GID-178: sync\.Mutex встроен в структуру — храните мьютекс именованным полем \(mu sync\.Mutex\), иначе Lock/Unlock попадают в API типа`
	data       map[string]string
}

// Встроенный указатель *sync.RWMutex.
type Registry struct {
	*sync.RWMutex // want `GID-178: sync\.RWMutex встроен в структуру — храните мьютекс именованным полем \(mu sync\.Mutex\), иначе Lock/Unlock попадают в API типа`
	items         []int
}

// Встраивание через алиас импорта пакета sync — детект по типу, не по тексту.
type Aliased struct {
	syncalias.Mutex // want `GID-178: sync\.Mutex встроен в структуру — храните мьютекс именованным полем \(mu sync\.Mutex\), иначе Lock/Unlock попадают в API типа`
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
