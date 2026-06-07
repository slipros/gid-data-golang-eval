// Позитив: model не зависит от dal.
package model

import "svc/dal/entity" // want `GID-132: package "svc/domain/model" must not import "svc/dal/entity"\. Fix: model does not depend on the dal layer`

type Legacy struct {
	Raw entity.Snapshot
}
