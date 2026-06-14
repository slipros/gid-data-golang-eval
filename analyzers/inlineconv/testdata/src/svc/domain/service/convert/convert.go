// Class 2 (negative) — the service's convert package: conversion lives here,
// inline entity literals are allowed.
package convert

import (
	"svc/dal/entity"
	"svc/domain/model"
)

func EntityCreateSnapshotFromModel(in model.CreateSnapshot) entity.CreateSnapshot {
	return entity.CreateSnapshot{Name: in.Name}
}

func ModelSnapshotFromEntity(in entity.Snapshot) model.Snapshot {
	return model.Snapshot{ID: in.ID, Name: in.Name}
}

func EntitySnapshotsFromModel() entity.Snapshots {
	return entity.Snapshots{entity.Snapshot{ID: "a"}}
}
