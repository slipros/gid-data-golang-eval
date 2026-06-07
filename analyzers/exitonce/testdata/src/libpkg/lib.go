// Класс «позитив»: библиотечный (не main) пакет.
// log.Fatal и logrus.Fatal* в любом месте не-main пакета запрещены.
package libpkg

import (
	stdlog "log"

	"github.com/sirupsen/logrus"
)

func mustLoad() {
	stdlog.Fatal("cannot load") // want `GID-181: log\.Fatal вне func main запрещён — верните error наверх`
}

func mustParse() {
	logrus.Fatalf("bad config") // want `GID-181: logrus\.Fatalf вне func main запрещён — верните error наверх`
}

// Метод logrus-логгера тоже считается exit-вызовом.
func withLogger(l *logrus.Logger) {
	l.Fatal("boom") // want `GID-181: logrus\.Fatal вне func main запрещён — верните error наверх`
}

// --- Негативный кейс: возврат error вместо завершения процесса ---

func load() error {
	return nil
}
