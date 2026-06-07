// Stub gderror для eval: конструктор обработки отсутствующего значения enum.
package errors

type UnhandledValueError struct {
	Value any
}

func (e *UnhandledValueError) Error() string { return "unhandled value" }

func NewUnhandledValueError(value any) error {
	return &UnhandledValueError{Value: value}
}
