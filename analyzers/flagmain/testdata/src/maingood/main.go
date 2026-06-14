// "Negative" class: in main the flag names are in snake_case, flag.Parse is allowed.
package main

import (
	"flag"
	"time"
)

var (
	maxRetries  = flag.Int("max_retries", 3, "retries")
	addr        = flag.String("addr", ":8080", "addr")
	readTimeout = flag.Duration("read_timeout_5s", 5*time.Second, "t")
)

func main() {
	var port int
	flag.IntVar(&port, "http_port", 8080, "port")
	flag.Bool("verbose", false, "verbose")

	flag.Parse()
	_ = maxRetries
	_ = addr
	_ = readTimeout
	_ = port
}
