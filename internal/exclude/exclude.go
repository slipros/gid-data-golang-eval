// Package exclude — централизованные исключения правил через settings
// линтера в .golangci.yml. Запись списка: "Метод" (любой тип)
// или "Тип.Метод" (конкретный тип).
package exclude

// Match сообщает, числится ли метод recvType.method в списке исключений.
func Match(list []string, recvType, method string) bool {
	for _, e := range list {
		if e == method || e == recvType+"."+method {
			return true
		}
	}
	return false
}
