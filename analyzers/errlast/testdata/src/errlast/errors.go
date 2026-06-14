// Boundary case: error-constructor functions in errors.go legitimately
// return a concrete error type — check 2 is not applied (by file).
package errlast

// newMyError — a private error constructor in errors.go: a concrete type is ok.
func newMyError(msg string) *MyError {
	return &MyError{msg: msg}
}

// NewMyError — an exported constructor: a concrete type in errors.go is ok.
func NewMyError() *MyError {
	return &MyError{}
}

// Check 1 (error is not last) applies in errors.go too.
func badOrder() (error, int) { // want `GID-190: error must be the last return value\. Fix: move it to the end`
	return nil, 0
}
