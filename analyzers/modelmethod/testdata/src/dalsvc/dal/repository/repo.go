// Неприменимость: /dal/repository вне scope GID-195.
package repository

import "dalsvc/domain/model"

func snapshotKey(s *model.Snapshot) string {
	return s.ID
}
