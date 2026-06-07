// Класс «позитив»: библиотечный (не main) пакет.
// log.Fatal и logrus.Fatal* в любом месте не-main пакета запрещены.
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

// Метод logrus-логгера тоже считается exit-вызовом.
func withLogger(l *logrus.Logger) {
	l.Fatal("boom") // want `GID-181: logrus\.Fatal is forbidden outside func main\. Fix: return an error up the call stack`
}

// --- Негативный кейс: возврат error вместо завершения процесса ---

func load() error {
	return nil
}
