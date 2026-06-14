// Eval for GID-161 (no panic outside main).
package nopanic

import "errors"

// --- Positive cases ---

func bad() {
	panic("boom") // want `GID-161: panic is allowed only in package main\. Fix: return an error instead`
}

// Boundary case: panic with an error argument.
func badErr(err error) {
	panic(err) // want `GID-161: panic is allowed only in package main\. Fix: return an error instead`
}

// --- Negative cases ---

func good() error {
	return errors.New("boom") //nolint // (GID-146 is checked by another linter)
}

// Boundary case: a local function named panic is not the builtin panic.
func shadowed() {
	panic := func(s string) {}
	panic("ok")
}
