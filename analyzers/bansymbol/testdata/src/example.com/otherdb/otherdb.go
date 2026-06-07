// Stub стороннего пакета с символом TQuery, но из ДРУГОГО пакета —
// он не должен флагаться (имя совпадает, путь — нет).
package otherdb

// TQuery — одноимённая функция другого пакета.
func TQuery[T any](query string) (T, error) {
	var zero T
	return zero, nil
}
