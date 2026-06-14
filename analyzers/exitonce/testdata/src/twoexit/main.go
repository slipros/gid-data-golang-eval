// "Positive" class: two os.Exit in func main — the second (duplicate) one is forbidden.
package main

import "os"

func main() {
	if len(someArgs()) == 0 {
		os.Exit(2) // the first call — acceptable
	}
	os.Exit(0) // want `GID-181: duplicate os\.Exit in main\. Fix: exit the program in a single place`
}

func someArgs() []string {
	return nil
}
