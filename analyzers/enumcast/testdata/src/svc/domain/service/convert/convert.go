// Eval GID-233: a direct cast between enums from different packages is
// forbidden; the conversion goes via a map with comma-ok and
// gderror.NewUnhandledValueError (GID-143). Fixtures intentionally violate
// the rule.
package convert

import (
	entityenum "svc/dal/entity/enum"
	modelenum "svc/domain/model/enum"

	gderror "gitlab.gid.team/gid-data/tech/golang/libs/helper.git/errors"
)

// --- Class 1: positive — cross-package enum→enum direct cast ---

// Direct cast entity enum → model enum: an unknown value silently crosses the boundary.
func ModelStatusCast(s entityenum.Status) modelenum.Status {
	return modelenum.Status(s) // want `GID-233: direct cast between enum types crosses a layer boundary unchecked\. Fix: convert via map with comma-ok \+ gderror\.NewUnhandledValueError \(see GID-143\)`
}

// Reverse direction (model enum → entity enum) is caught too.
func EntityStatusCast(s modelenum.Status) entityenum.Status {
	v := entityenum.Status(s) // want `GID-233: direct cast between enum types crosses a layer boundary unchecked\. Fix: convert via map with comma-ok \+ gderror\.NewUnhandledValueError \(see GID-143\)`
	return v
}

// A typed const of a foreign enum is still a foreign enum value (see .feature decision).
func DefaultEntityStatusCast() modelenum.Status {
	return modelenum.Status(entityenum.StatusActive) // want `GID-233: direct cast between enum types crosses a layer boundary unchecked\. Fix: convert via map with comma-ok \+ gderror\.NewUnhandledValueError \(see GID-143\)`
}

// --- Class 2: negative — the GID-143 map converter stays clean ---

// statusFromEntity is the canonical converter: no cast anywhere.
var statusFromEntity = map[entityenum.Status]modelenum.Status{
	entityenum.StatusActive:  modelenum.StatusActive,
	entityenum.StatusBlocked: modelenum.StatusBlocked,
}

// ModelStatusFromEntity converts via the map with comma-ok — clean.
func ModelStatusFromEntity(s entityenum.Status) (modelenum.Status, error) {
	v, ok := statusFromEntity[s]
	if !ok {
		return "", gderror.NewUnhandledValueError(s)
	}
	return v, nil
}

// --- Class 3: boundary — close to the line, but allowed ---

// Cast from plain string is allowed.
func ModelStatusFromString(s string) modelenum.Status {
	return modelenum.Status(s)
}

// Cast to plain string is allowed.
func StringFromEntityStatus(s entityenum.Status) string {
	return string(s)
}

// Cast of an untyped constant/literal is allowed.
func DefaultModelStatus() modelenum.Status {
	return modelenum.Status("active")
}

// --- Class 4: non-applicability — named string types without consts ---

// Neither Label nor Raw has typed consts — they are not enums.
func LabelFromRaw(r entityenum.Raw) modelenum.Label {
	return modelenum.Label(r)
}

// Destination Raw has no typed consts — not an enum, even though the source is.
func RawFromStatus(s entityenum.Status) entityenum.Raw {
	return entityenum.Raw(s)
}
