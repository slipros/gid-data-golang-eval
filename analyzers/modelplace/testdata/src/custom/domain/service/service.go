// Eval for GID-270 (model-place) with custom settings: settings.suffixes
// replaces the options-suffix defaults, settings.exclude exempts a struct
// (the "Struct" form) in both parts.
package service

// TriggerSpec — boundary: with settings.suffixes: ["Spec"] the type is
// settings, not a data model.
type TriggerSpec struct {
	Retry int
}

// LegacyDTO — non-applicability: settings.exclude exempts the whole struct.
type LegacyDTO struct {
	Code string
}

// Plain — positive: no suffix from the custom list, not excluded.
type Plain struct { // want `GID-270: data struct "Plain" is declared in /domain/service`
	Code string
}

// LegacyOptions — positive: custom suffixes REPLACE the defaults, so the
// Options suffix no longer exempts.
type LegacyOptions struct { // want `GID-270: data struct "LegacyOptions" is declared in /domain/service`
	Batch int
}

// --- Part C with custom settings ---

// makeLegacy — non-applicability: settings.exclude exempts the type on the
// return as well as at its declaration.
func makeLegacy() LegacyDTO {
	return LegacyDTO{}
}

// runSpec — an unexported struct with the custom "Spec" suffix: settings.suffixes
// exempts it in part C too.
type runSpec struct {
	retry int
}

func makeSpec() runSpec {
	return runSpec{}
}
