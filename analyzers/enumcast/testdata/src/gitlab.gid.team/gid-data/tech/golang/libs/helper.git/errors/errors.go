// Stub gderror for eval: constructor handling an unhandled enum value.
package errors

type UnhandledValueError struct {
	Value any
}

func (e *UnhandledValueError) Error() string { return "unhandled value" }

func NewUnhandledValueError(value any) error {
	return &UnhandledValueError{Value: value}
}
