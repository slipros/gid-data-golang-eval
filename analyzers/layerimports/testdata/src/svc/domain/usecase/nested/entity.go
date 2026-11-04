package nested

import "svc/dal/entity"       // want `GID-132:.*usecase works only with model`
func Entity() entity.Snapshot { return entity.Snapshot{} }
