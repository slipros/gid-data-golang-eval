// "Positive + negative" class: package main.
// A helper with os.Exit outside main is forbidden; a single os.Exit at the end of main is fine.
package main

import "os"

// --- Positive case: an exit call outside func main (in a helper) ---

func fail() {
	os.Exit(1) // want `GID-181: os\.Exit is forbidden outside func main\. Fix: return an error up the call stack`
}

// --- Negative case: exactly one os.Exit at the end of main ---

func run() error {
	return nil
}

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}
