// Not applicable: the convert subpackage is not the root of the service layer.
package convert

import "svc/domain/model"

func snapshotName(s *model.Snapshot) string {
	return s.Name
}
