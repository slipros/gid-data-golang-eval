// "Boundary" class: a local package named flag is NOT the stdlib "flag",
// the import path differs, so the rule does not fire even in a library.
package ownflag

import "ownflag/flag"

func use() {
	flag.String("maxRetries") // the name is not snake_case, but this is a foreign flag — no diagnostic
	flag.Parse()
}
