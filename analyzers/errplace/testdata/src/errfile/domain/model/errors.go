// Negative: errors.go is an allowed file, error declarations here are fine.
package model

import "github.com/pkg/errors"

var ErrSnapshotNotFound = errors.New("snapshot not found")

var ErrSnapshotExpired = errors.New("snapshot expired")
