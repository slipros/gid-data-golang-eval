// Eval GID-180: not applicable — a package without init().
package noinit

import "os"

// No func init() at all → the rule is not activated,
// even with a go statement and I/O calls present in ordinary functions.

func Run() {
	go func() {}()
	_, _ = os.Open("/etc/hosts")
	db := os.Getenv("DB")
	_ = db
}
