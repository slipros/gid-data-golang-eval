// Negative: the canonical order import -> const -> var -> types -> functions.
package constvarorder

import "time"

const (
	defaultPartSize = 5 * 1024
)

const singleConst = "ok"

var DefaultTimeout = 5 * time.Second

type Hello struct{}

// Not applicable: const/var inside a function — the file order does not apply.
func inner() time.Duration {
	const localConst = 2
	var localVar = time.Second * localConst
	return localVar
}
