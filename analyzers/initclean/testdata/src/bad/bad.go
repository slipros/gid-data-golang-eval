// Eval GID-180: positive and boundary violations in init().
package bad

import (
	"database/sql"
	"os"
)

// Positive: starting a goroutine directly in init.
func init() {
	go func() {} () // want `GID-180: a goroutine in init\(\) is forbidden`
}

// Positive: an I/O call os.Open in init.
func init() {
	f, _ := os.Open("/etc/hosts") // want `GID-180: an I/O call os\.Open in init\(\) is forbidden`
	_ = f
}

// Positive: an I/O call sql.Open in init.
func init() {
	db, _ := sql.Open("postgres", "") // want `GID-180: an I/O call database/sql\.Open in init\(\) is forbidden`
	_ = db
}

// Boundary: a closure is declared and called in init, with os.Open inside — matched,
// because the closure's body is walked as part of init.
func init() {
	fn := func() {
		_, _ = os.Open("/tmp/x") // want `GID-180: an I/O call os\.Open in init\(\) is forbidden`
	}
	fn()
}

// Boundary: a goroutine in a nested block of init — matched.
func init() {
	{
		go func() {}() // want `GID-180: a goroutine in init\(\) is forbidden`
	}
}
