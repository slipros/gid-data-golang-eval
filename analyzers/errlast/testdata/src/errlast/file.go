// Eval for GID-190 (error is the last result; concrete error types are forbidden).
package errlast

// MyError — a concrete named type implementing error (via a pointer).
type MyError struct{ msg string }

func (e *MyError) Error() string { return e.msg }

// ValError — a concrete named type implementing error by value.
type ValError struct{ code int }

func (e ValError) Error() string { return "val error" }

// T — an ordinary struct, does not implement error.
type T struct{ Name string }

// ErrIface — a custom error interface (extends error). A deliberate decision.
type ErrIface interface {
	error
	Code() int
}

// --- Class 1: positive (violations) ---

// error is not last — an int follows it.
func f() (error, int) { // want `GID-190: error must be the last return value\. Fix: move it to the end`
	return nil, 0
}

// the result is a concrete error type (*MyError), not the error interface.
func g() *MyError { // want `GID-190: return the error interface, not \*errlast.MyError\. Fix: a concrete type in the error position causes a typed-nil trap`
	return nil
}

// method: error is not last (there is ok after it).
func (t T) Do() (err error, ok bool) { // want `GID-190: error must be the last return value\. Fix: move it to the end`
	return nil, false
}

// the result is a concrete error type by value (ValError).
func valErr() ValError { // want `GID-190: return the error interface, not errlast.ValError\. Fix: a concrete type in the error position causes a typed-nil trap`
	return ValError{}
}

// --- Class 2: negative (clean code) ---

// error is last — normal.
func ok1() (int, error) {
	return 0, nil
}

// (T, error) where T is an ordinary struct, error is last — normal.
func ok2() (T, error) {
	return T{}, nil
}

// a single error result — normal.
func e() error {
	return nil
}

// no error among the results — non-applicability.
func plain() (int, string) {
	return 0, ""
}

// --- Class 3: boundary ---

// the result is a custom error interface ErrIface (extends error) — NOT matched.
func h() ErrIface {
	return nil
}

// a single result (error) — ok.
func single() error {
	return nil
}

// several results, error is last, among the others — a concrete non-error type.
func ok3() (T, ValError, error) { // want `GID-190: return the error interface, not errlast.ValError\. Fix: a concrete type in the error position causes a typed-nil trap`
	return T{}, ValError{}, nil
}
