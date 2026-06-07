// Граничный кейс: функции-конструкторы ошибок в errors.go легитимно
// возвращают конкретный error-тип — проверка 2 не применяется (по файлу).
package errlast

// newMyError — приватный конструктор ошибки в errors.go: конкретный тип ок.
func newMyError(msg string) *MyError {
	return &MyError{msg: msg}
}

// NewMyError — экспортируемый конструктор: конкретный тип в errors.go ок.
func NewMyError() *MyError {
	return &MyError{}
}

// Проверка 1 (error не последний) действует и в errors.go.
func badOrder() (error, int) { // want `GID-190: error must be the last return value\. Fix: move it to the end`
	return nil, 0
}
