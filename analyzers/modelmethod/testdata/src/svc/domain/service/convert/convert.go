// Неприменимость: convert-подпакет не является корнем service-слоя.
package convert

import "svc/domain/model"

func snapshotName(s *model.Snapshot) string {
	return s.Name
}
