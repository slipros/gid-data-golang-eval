// Stub of a third-party package with a TQuery symbol, but from a DIFFERENT package —
// it must not be flagged (the name matches, the path does not).
package otherdb

// TQuery — a same-named function from another package.
func TQuery[T any](query string) (T, error) {
	var zero T
	return zero, nil
}
