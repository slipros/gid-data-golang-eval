// Not applicable: /dal/repository is outside the GID-195 scope.
package repository

import "dalsvc/domain/model"

func snapshotKey(s *model.Snapshot) string {
	return s.ID
}
