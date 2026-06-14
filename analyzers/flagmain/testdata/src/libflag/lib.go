// "Positive" class: registering a flag in a library (non-main) package is forbidden.
package libflag

import "flag"

var maxRetries = flag.Int("max_retries", 3, "retries") // want `GID-192: registering a flag outside package main is forbidden\. Fix: declare flags in the binary, let libraries take parameters`

func register() {
	flag.String("addr", ":8080", "listen addr") // want `GID-192: registering a flag outside package main is forbidden\. Fix: declare flags in the binary, let libraries take parameters`
}

// Boundary: the flag name is dynamic — part 2 is not evaluated, but part 1
// (registration outside main) still fires.
func registerDynamic(name string) {
	flag.String(name, "", "addr") // want `GID-192: registering a flag outside package main is forbidden\. Fix: declare flags in the binary, let libraries take parameters`
}
