// Класс «неприменимость»: библиотечный пакет без exit-вызовов —
// правило не срабатывает.
package cleanlib

import (
	stdlog "log"

	"github.com/sirupsen/logrus"
)

// Используем нефатальные методы — они не подпадают под GID-181.
func process(l *logrus.Entry) error {
	l.Info("processing")
	stdlog.Print("processing")
	return nil
}

func reportErr(l *logrus.Entry, err error) error {
	l.WithError(err).Error("failed")
	return err
}
