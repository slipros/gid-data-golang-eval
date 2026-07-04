// Positive (GID-235): a convert package logs — a converter must be a pure
// function without side effects.
package convert

import (
	"github.com/sirupsen/logrus" // want `GID-235: convert package "svc/server/grpc/handler/convert" must not import "github.com/sirupsen/logrus" — a converter is a pure function over vocabulary types \(model/entity/dto/client/pb\); business logic and side effects live in their layers`
)

var log = &logrus.Logger{}

func LogAndConvert(id string) string {
	log.Info(id)
	return id
}
