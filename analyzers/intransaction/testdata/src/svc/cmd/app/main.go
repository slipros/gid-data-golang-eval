// Eval for GID-175: check 3 does not apply outside service/usecase.
// An anonymous tx-signature in main (not a service/usecase layer) — check 3
// does not apply, no diagnostic. Check 1 catches only named type
// declarations; here the type is anonymous in a parameter, so it is clean.
package main

import "context"

func run(tx func(ctx context.Context, fn func(ctx context.Context) error) error) {
	_ = tx
}

func main() {}
