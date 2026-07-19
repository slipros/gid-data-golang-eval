// Stub of the errors helper on the nomilk module path (not gitlab): the same
// NewUnhandledValueError constructor lives under a different import path. Used
// to prove GID-143 matches the handler by symbol name, not by module path.
package ehelper

type UnhandledValueError struct {
	Value any
}

func (e *UnhandledValueError) Error() string { return "unhandled value" }

func NewUnhandledValueError(value any) error {
	return &UnhandledValueError{Value: value}
}
