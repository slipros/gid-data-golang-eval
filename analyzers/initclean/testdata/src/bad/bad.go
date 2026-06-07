// Eval GID-180: позитивные и граничные нарушения в init().
package bad

import (
	"database/sql"
	"os"
)

// Позитив: запуск горутины прямо в init.
func init() {
	go func() {} () // want `GID-180: a goroutine in init\(\) is forbidden`
}

// Позитив: I/O-вызов os.Open в init.
func init() {
	f, _ := os.Open("/etc/hosts") // want `GID-180: an I/O call os\.Open in init\(\) is forbidden`
	_ = f
}

// Позитив: I/O-вызов sql.Open в init.
func init() {
	db, _ := sql.Open("postgres", "") // want `GID-180: an I/O call database/sql\.Open in init\(\) is forbidden`
	_ = db
}

// Граничный: замыкание объявлено и вызвано в init, внутри os.Open — матчится,
// так как тело замыкания обходится как часть init.
func init() {
	fn := func() {
		_, _ = os.Open("/tmp/x") // want `GID-180: an I/O call os\.Open in init\(\) is forbidden`
	}
	fn()
}

// Граничный: горутина во вложенном блоке init — матчится.
func init() {
	{
		go func() {}() // want `GID-180: a goroutine in init\(\) is forbidden`
	}
}
