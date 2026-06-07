// Stub gderror для eval: типизированные ошибки внутренней библиотеки
// разрешены в любом слое.
package errors

type UnhandledValueError struct {
	Value any
}

func (e *UnhandledValueError) Error() string { return "unhandled value" }

func NewUnhandledValueError(value any) error {
	return &UnhandledValueError{Value: value}
}
