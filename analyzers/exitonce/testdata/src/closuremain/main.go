// "Boundary" class: os.Exit inside a closure in main counts as a call in main
// (not "outside main"). A single such call — no diagnostic.
package main

import "os"

func main() {
	done := func(code int) {
		os.Exit(code)
	}
	done(0)
}
