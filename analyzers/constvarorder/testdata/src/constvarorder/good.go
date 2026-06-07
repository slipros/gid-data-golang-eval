// Негатив: канонический порядок import -> const -> var -> типы -> функции.
package constvarorder

import "time"

const (
	defaultPartSize = 5 * 1024
)

const singleConst = "ok"

var DefaultTimeout = 5 * time.Second

type Hello struct{}

// Неприменимость: const/var внутри функции — порядок файла не касается.
func inner() time.Duration {
	const localConst = 2
	var localVar = time.Second * localConst
	return localVar
}
