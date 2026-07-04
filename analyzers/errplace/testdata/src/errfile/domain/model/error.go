// Negative: error.go is the (sole, default) allowed file — declarations here are fine.
package model

import "github.com/pkg/errors"

var ErrSnapshotArchived = errors.New("snapshot archived")

var ErrSnapshotDeleted = errors.New("snapshot deleted")
