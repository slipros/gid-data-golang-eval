// Eval for GID-270 (model-place), part A with custom settings:
// settings.exclude (the "Struct" form) applies to a convert package too.
package convert

// Exempt — non-applicability: excluded by settings.exclude.
type Exempt struct {
	Code string
}

// Extra — positive: not on the exclusion list.
type Extra struct { // want `GID-270: type "Extra" is declared in a convert package`
	MessageID string
	Body      string
}
