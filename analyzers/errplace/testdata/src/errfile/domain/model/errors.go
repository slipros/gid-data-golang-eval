// Positive: errors.go is no longer an allowed file by default —
// only error.go is; declaring errors here now violates GID-169.
package model

import "github.com/pkg/errors"

var ErrSnapshotNotFound = errors.New("snapshot not found") // want `GID-169: error "ErrSnapshotNotFound" is declared in errors\.go\. Fix: keep layer errors in error\.go`

var ErrSnapshotExpired = errors.New("snapshot expired") // want `GID-169: error "ErrSnapshotExpired" is declared in errors\.go\. Fix: keep layer errors in error\.go`
