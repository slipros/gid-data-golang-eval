// "Inapplicability" class: a library package without exit calls —
// the rule does not fire.
package cleanlib

import (
	stdlog "log"

	"github.com/sirupsen/logrus"
)

// We use non-fatal methods — they do not fall under GID-181.
func process(l *logrus.Entry) error {
	l.Info("processing")
	stdlog.Print("processing")
	return nil
}

func reportErr(l *logrus.Entry, err error) error {
	l.WithError(err).Error("failed")
	return err
}
