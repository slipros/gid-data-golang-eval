// Класс «позитив»: в main имя флага не в snake_case.
package main

import (
	"flag"
	"time"
)

var (
	maxRetries = flag.String("maxRetries", "3", "retries") // want `GID-192: имя флага "maxRetries" — используйте snake_case`
	maxRetry   = flag.Int("max-retries", 3, "retries")     // want `GID-192: имя флага "max-retries" — используйте snake_case`
)

func main() {
	flag.Bool("DryRun", false, "dry run")           // want `GID-192: имя флага "DryRun" — используйте snake_case`
	flag.Duration("read-timeout", time.Second, "t") // want `GID-192: имя флага "read-timeout" — используйте snake_case`

	var port int
	flag.IntVar(&port, "httpPort", 8080, "port") // want `GID-192: имя флага "httpPort" — используйте snake_case`

	flag.Parse()
	_ = maxRetries
	_ = maxRetry
}
