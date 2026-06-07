// Неприменимость: в пакете нет структур — правилу нечего проверять.
package nostructs

import "sync"

// Функция со встроенным мьютексом? Нет — мьютекс лежит именованной переменной.
func New() *sync.Mutex {
	var mu sync.Mutex
	return &mu
}

// Интерфейс — встраивание мьютекса в интерфейсы не бывает, и тут его нет.
type Locker interface {
	Lock()
	Unlock()
}
