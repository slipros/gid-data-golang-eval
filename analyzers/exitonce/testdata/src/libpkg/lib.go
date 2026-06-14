// "Positive" class: a library (non-main) package.
// log.Fatal and logrus.Fatal* are forbidden anywhere in a non-main package.
package libpkg

import (
	stdlog "log"

	"github.com/sirupsen/logrus"
)

func mustLoad() {
	stdlog.Fatal("cannot load") // want `GID-181: log\.Fatal is forbidden outside func main\. Fix: return an error up the call stack`
}

func mustParse() {
	logrus.Fatalf("bad config") // want `GID-181: logrus\.Fatalf is forbidden outside func main\. Fix: return an error up the call stack`
}

// A logrus logger method also counts as an exit call.
func withLogger(l *logrus.Logger) {
	l.Fatal("boom") // want `GID-181: logrus\.Fatal is forbidden outside func main\. Fix: return an error up the call stack`
}

// --- Negative case: returning an error instead of terminating the process ---

func load() error {
	return nil
}
