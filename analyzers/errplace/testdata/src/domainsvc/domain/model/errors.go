// Negative: /domain/model is the home of domain errors, declaration is allowed here.
// Creation — via github.com/pkg/errors (GID-146).
package model

import "github.com/pkg/errors"

var ErrSnapshotNotFound = errors.New("snapshot not found")

var ErrSnapshotExpired = errors.New("snapshot expired")
