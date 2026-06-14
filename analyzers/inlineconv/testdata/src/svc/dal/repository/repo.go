// Class 4 (not applicable) — /dal/repository is not part of the domain layer,
// the rule does not apply: entity literals here are not flagged.
package repository

import "svc/dal/entity"

func build(id string) entity.Snapshot {
	return entity.Snapshot{ID: id}
}

func buildSlice() entity.Snapshots {
	return entity.Snapshots{{ID: "a"}}
}
