// Stub gderror: конструктор сам собирает стек (исключение GID-177).
package gderror

type unhandledValueError struct{ v any }

func (e unhandledValueError) Error() string { return "unhandled value" }

func NewUnhandledValueError(v any) error {
	return unhandledValueError{v: v} // want `GID-177: a static error is returned without a stack`
}
