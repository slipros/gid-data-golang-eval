// Package withtests — boundary fixture of GID-262: settings.include-tests
// judges _test.go files too (the requirement map has been extracted already).
package withtests

// Reason returns the reason code of the last record.
func Reason() string { return "ok" }
