// Eval для GID-195: usecase — тоже scope правила.
package usecase

import "svc/domain/model"

func snapshotKey(s *model.Snapshot) string { // want `GID-195: private function "snapshotKey" works only with the model.Snapshot value\. Fix: this is model behaviour, make it a public method of that type`
	return s.ID
}
