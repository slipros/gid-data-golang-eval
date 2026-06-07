// Eval GID-180: позитивные и граничные нарушения в init().
package bad

import (
	"database/sql"
	"os"
)

// Позитив: запуск горутины прямо в init.
func init() {
	go func() {} () // want `GID-180: goroutine в init\(\) запрещена`
}

// Позитив: I/O-вызов os.Open в init.
func init() {
	f, _ := os.Open("/etc/hosts") // want `GID-180: I/O-вызов os\.Open в init\(\) запрещён`
	_ = f
}

// Позитив: I/O-вызов sql.Open в init.
func init() {
	db, _ := sql.Open("postgres", "") // want `GID-180: I/O-вызов database/sql\.Open в init\(\) запрещён`
	_ = db
}

// Граничный: замыкание объявлено и вызвано в init, внутри os.Open — матчится,
// так как тело замыкания обходится как часть init.
func init() {
	fn := func() {
		_, _ = os.Open("/tmp/x") // want `GID-180: I/O-вызов os\.Open в init\(\) запрещён`
	}
	fn()
}

// Граничный: горутина во вложенном блоке init — матчится.
func init() {
	{
		go func() {}() // want `GID-180: goroutine в init\(\) запрещена`
	}
}
