// "Boundary" class: defer + one os.Exit in main is exactly one call, fine.
package main

import "os"

func cleanup() {}

func main() {
	defer cleanup()
	os.Exit(0)
}
