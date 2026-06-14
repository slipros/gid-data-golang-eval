// Package model — fixture for the settings:
// settings.names = ["ErrOops"], settings.exclude = ["ErrLegacy"].
package model

import "errors"

// Positive: a name from the custom names list.
var ErrOops = errors.New("oops") // want `GID-234: generic error name "ErrOops" in domain model\. Fix: bind it to the entity: ErrSnapshotOops`

// Boundary: the default list is replaced — ErrNotFound is no longer forbidden.
var ErrNotFound = errors.New("not found")

// Boundary: ErrLegacy is in names but excluded via settings.exclude.
var ErrLegacy = errors.New("legacy")
