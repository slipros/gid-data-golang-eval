// Eval GID-214: negative — package main, creating a logger is allowed.
package main

import (
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()      // composition root (main) — ok
	_ = logrus.StandardLogger() // composition root (main) — ok
	_ = logger
}
