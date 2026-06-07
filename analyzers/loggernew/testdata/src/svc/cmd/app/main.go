// Eval GID-214: негатив — package main, создание логгера разрешено.
package main

import (
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()      // composition root (main) — ок
	_ = logrus.StandardLogger() // composition root (main) — ок
	_ = logger
}
