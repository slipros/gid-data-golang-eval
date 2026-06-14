// Class 4 (not applicable) — _test.go is skipped: an inline entity literal
// in a test is not flagged.
package service

import "svc/dal/entity"

func buildTestSnapshot(id string) entity.Snapshot {
	return entity.Snapshot{ID: id}
}
