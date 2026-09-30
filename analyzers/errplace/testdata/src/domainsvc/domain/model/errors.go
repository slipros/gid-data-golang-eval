// Negative: /domain/model is the home of domain errors, declaration is allowed here.
// Creation — via github.com/pkg/errors (GID-146).
package model

import "github.com/pkg/errors"

var ErrSnapshotNotFound = errors.New("snapshot not found")

var ErrSnapshotExpired = errors.New("snapshot expired")

// Negative: an error TYPE is allowed in model — it is the layer's error home too.
type ValidationError struct{ Field string }

func (v ValidationError) Error() string { return "invalid " + v.Field }
