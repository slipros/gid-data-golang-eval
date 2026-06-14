// Eval GID-233: _test.go files are skipped — the cast below yields no diagnostic.
package convert

import (
	entityenum "svc/dal/entity/enum"
	modelenum "svc/domain/model/enum"
)

// castInTest would violate GID-233 in regular code, but _test.go is skipped.
func castInTest(s entityenum.Status) modelenum.Status {
	return modelenum.Status(s)
}

var _ = castInTest
