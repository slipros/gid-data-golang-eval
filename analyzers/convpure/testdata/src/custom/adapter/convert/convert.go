// Eval for settings.packages: a custom in-house library replaces the
// default logrus ban.
package convert

import (
	somelib "example.com/inhouse/somelib" // want `GID-235: convert package "custom/adapter/convert" must not import "example.com/inhouse/somelib" — a converter is a pure function over vocabulary types\. Fix: import only model/entity/dto/client/pb; move the logic or side effect to its layer and pass the result into the converter`

	"github.com/sirupsen/logrus" // the default list is replaced — not flagged
)

var log = &logrus.Logger{}

func Convert(v somelib.Value) string {
	log.Info(v)
	return v.String()
}
