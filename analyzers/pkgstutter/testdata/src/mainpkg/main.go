// Eval for GID-193: not applicable — package main is not checked (bootstrap).
package main

// MainOptions would be a violation in an ordinary package, but package main is excluded.
type MainOptions struct {
	Addr string
}

func main() {}
