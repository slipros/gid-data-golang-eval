// "Boundary" class: the flag name is not a string constant —
// part 2 (snake_case) is not checked (the name is unknown statically).
// Package main, so part 1 does not apply.
package main

import "flag"

func name() string { return "maxRetries" }

func main() {
	dyn := name()
	flag.String(dyn, "3", "retries") // the name is dynamic — no diagnostic
	flag.Parse()
}
