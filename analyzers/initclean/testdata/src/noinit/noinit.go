// Eval GID-180: неприменимость — пакет без init().
package noinit

import "os"

// Нет ни одной func init() → правило не активируется,
// даже при наличии go-statement и I/O-вызовов в обычных функциях.

func Run() {
	go func() {}()
	_, _ = os.Open("/etc/hosts")
	db := os.Getenv("DB")
	_ = db
}
