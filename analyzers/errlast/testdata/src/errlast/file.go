// Eval для GID-190 (error — последний результат; конкретные error-типы запрещены).
package errlast

// MyError — конкретный именованный тип, реализующий error (через указатель).
type MyError struct{ msg string }

func (e *MyError) Error() string { return e.msg }

// ValError — конкретный именованный тип, реализующий error по значению.
type ValError struct{ code int }

func (e ValError) Error() string { return "val error" }

// T — обычная структура, не реализует error.
type T struct{ Name string }

// ErrIface — кастомный error-интерфейс (расширяет error). Осознанное решение.
type ErrIface interface {
	error
	Code() int
}

// --- Класс 1: позитивные (нарушения) ---

// error не последний — после него идёт int.
func f() (error, int) { // want `GID-190: error must be the last return value\. Fix: move it to the end`
	return nil, 0
}

// результат — конкретный error-тип (*MyError), а не интерфейс error.
func g() *MyError { // want `GID-190: return the error interface, not \*errlast.MyError\. Fix: a concrete type in the error position causes a typed-nil trap`
	return nil
}

// метод: error не последний (есть ok после него).
func (t T) Do() (err error, ok bool) { // want `GID-190: error must be the last return value\. Fix: move it to the end`
	return nil, false
}

// результат — конкретный error-тип по значению (ValError).
func valErr() ValError { // want `GID-190: return the error interface, not errlast.ValError\. Fix: a concrete type in the error position causes a typed-nil trap`
	return ValError{}
}

// --- Класс 2: негативные (чистый код) ---

// error последний — норма.
func ok1() (int, error) {
	return 0, nil
}

// (T, error) где T — обычная struct, error последний — норма.
func ok2() (T, error) {
	return T{}, nil
}

// единственный результат error — норма.
func e() error {
	return nil
}

// без error в результатах — неприменимость.
func plain() (int, string) {
	return 0, ""
}

// --- Класс 3: граничные ---

// результат — кастомный error-интерфейс ErrIface (расширяет error) — НЕ матчится.
func h() ErrIface {
	return nil
}

// единственный результат (error) — ок.
func single() error {
	return nil
}

// несколько результатов, error последний, среди прочих — конкретный тип не-error.
func ok3() (T, ValError, error) { // want `GID-190: return the error interface, not errlast.ValError\. Fix: a concrete type in the error position causes a typed-nil trap`
	return T{}, ValError{}, nil
}
