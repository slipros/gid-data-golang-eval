// Eval GID-143: a map-based enum conversion handles a missing key
// via gderror.NewUnhandledValueError. Applies in a convert package.
package convert

import gderror "gitlab.gid.team/gid-data/tech/golang/libs/helper.git/errors"

// An enum per GID-123: a named type based on string.
type (
	EntityStatus string
	ModelStatus  string
)

// statusMap — an enum→enum conversion (both the key and the value are named).
var statusMap = map[EntityStatus]ModelStatus{
	"active": "active",
}

// --- Class 1: positive ---

// A single assignment (not comma-ok): a missing key silently yields the zero value.
func ModelStatusFromEntity(s EntityStatus) ModelStatus {
	v := statusMap[s] // want `GID-143: enum conversion via map without comma-ok\. Fix: a missing key must return gderror.NewUnhandledValueError`
	return v
}

// A use as an expression (not comma-ok) — also without handling.
func ModelStatusExpr(s EntityStatus) ModelStatus {
	return statusMap[s] // want `GID-143: enum conversion via map without comma-ok\. Fix: a missing key must return gderror.NewUnhandledValueError`
}

// comma-ok is present, but the function has NO NewUnhandledValueError call.
func ModelStatusNoHandler(s EntityStatus) ModelStatus {
	v, ok := statusMap[s] // want `GID-143: a missing enum-conversion key must be handled with gderror.NewUnhandledValueError`
	if !ok {
		return ""
	}
	return v
}

// --- Class 2: negative ---

// comma-ok + handling of the missing key via gderror.NewUnhandledValueError.
func ModelStatusFromEntityOK(s EntityStatus) (ModelStatus, error) {
	v, ok := statusMap[s]
	if !ok {
		return "", gderror.NewUnhandledValueError(s)
	}
	return v, nil
}

// --- Class 3: edge ---

// The map key is a basic string (not an enum) → not matched.
var titleMap = map[string]ModelStatus{"active": "active"}

func ModelTitleFromString(s string) ModelStatus {
	return titleMap[s]
}

// The map value is a basic type (not named) → not matched (this is not an enum conversion).
var weightMap = map[EntityStatus]int{"active": 1}

func WeightFromStatus(s EntityStatus) int {
	return weightMap[s]
}
