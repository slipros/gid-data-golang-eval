// "Positive" class: in main the flag name is not in snake_case.
package main

import (
	"flag"
	"time"
)

var (
	maxRetries = flag.String("maxRetries", "3", "retries") // want `GID-192: flag name "maxRetries"\. Fix: use snake_case`
	maxRetry   = flag.Int("max-retries", 3, "retries")     // want `GID-192: flag name "max-retries"\. Fix: use snake_case`
)

func main() {
	flag.Bool("DryRun", false, "dry run")           // want `GID-192: flag name "DryRun"\. Fix: use snake_case`
	flag.Duration("read-timeout", time.Second, "t") // want `GID-192: flag name "read-timeout"\. Fix: use snake_case`

	var port int
	flag.IntVar(&port, "httpPort", 8080, "port") // want `GID-192: flag name "httpPort"\. Fix: use snake_case`

	flag.Parse()
	_ = maxRetries
	_ = maxRetry
}
