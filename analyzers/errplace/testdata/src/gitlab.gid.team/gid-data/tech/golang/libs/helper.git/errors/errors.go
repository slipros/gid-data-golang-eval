// Stub of gderror for eval: typed errors of the internal library
// are allowed in any layer.
package errors

type UnhandledValueError struct {
	Value any
}

func (e *UnhandledValueError) Error() string { return "unhandled value" }

func NewUnhandledValueError(value any) error {
	return &UnhandledValueError{Value: value}
}
