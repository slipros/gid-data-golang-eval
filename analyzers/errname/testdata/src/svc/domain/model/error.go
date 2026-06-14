package model

import "errors"

// Positive: generic names in /domain/model.
var (
	ErrNotFound      = errors.New("not found")      // want `GID-234: generic error name "ErrNotFound" in domain model\. Fix: bind it to the entity: ErrSnapshotNotFound`
	ErrAlreadyExists = errors.New("already exists") // want `GID-234: generic error name "ErrAlreadyExists" in domain model\. Fix: bind it to the entity: ErrSnapshotAlreadyExists`
	ErrInternal      = errors.New("internal")       // want `GID-234: generic error name "ErrInternal" in domain model\. Fix: bind it to the entity: ErrSnapshotInternal`
)

// Positive: explicit error type without initialization.
var ErrConflict error // want `GID-234: generic error name "ErrConflict" in domain model\. Fix: bind it to the entity: ErrSnapshotConflict`

// Negative: names bound to the entity — a generic suffix is not forbidden.
var (
	ErrSnapshotNotFound      = errors.New("snapshot not found")
	ErrSnapshotAlreadyExists = errors.New("snapshot already exists")
	ErrSnapshotExpired       = errors.New("snapshot expired")
)
