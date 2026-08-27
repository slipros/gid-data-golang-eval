// Eval for GID-273 (closure-ctor): a function building and returning a closure
// is an inline constructor of it and takes the New/new prefix.
package usecase

import "svc/domain/model"

// WebhookTriggerV2 — the entity of the incident (consent-webhook-trigger).
type WebhookTriggerV2 struct {
	filter func(identifier *model.ConsentEventV2UserIdentifier) bool
	scopes []string
}

// --- Positive class ---

// contactFilter — the incident shape verbatim: the method assembles a
// predicate and hands it over.
func (w *WebhookTriggerV2) contactFilter( // want `GID-273: method WebhookTriggerV2\.contactFilter builds and returns a closure — an inline constructor of it, not a value of the layer\. Fix: rename it to newContactFilter \(GID-104\) \(or //nolint:gidclosurector when the name is fixed by a convention\)`
	scope model.WebhookTriggerV2ContactFilterScope,
) func(identifier *model.ConsentEventV2UserIdentifier) bool {
	return func(identifier *model.ConsentEventV2UserIdentifier) bool {
		return identifier.Type == string(scope)
	}
}

// statusFilter — the closure travels through a local variable: the same
// construction, one step apart.
func statusFilter(status string) func(string) bool { // want `GID-273: function statusFilter builds and returns a closure`
	match := func(candidate string) bool {
		return candidate == status
	}

	return match
}

// ScopePredicate — an exported builder: the fix asks for the exported prefix.
func ScopePredicate(scope string) model.Predicate { // want `GID-273: function ScopePredicate builds and returns a closure — an inline constructor of it, not a value of the layer\. Fix: rename it to NewScopePredicate`
	return func(identifier *model.ConsentEventV2UserIdentifier) bool {
		return identifier.Value == scope
	}
}

// branchFilter — boundary: the literal is returned from inside a branch.
func branchFilter(enabled bool) func() bool { // want `GID-273: function branchFilter builds and returns a closure`
	if enabled {
		return func() bool { return true }
	}

	return nil
}

// --- Negative class ---

// newContactFilter — the fix itself: the constructor prefix is there.
func newContactFilter(scope string) func(string) bool {
	return func(candidate string) bool { return candidate == scope }
}

// NewScopePredicate — the exported form of the same.
func NewScopePredicate(scope string) model.Predicate {
	return func(identifier *model.ConsentEventV2UserIdentifier) bool {
		return identifier.Value == scope
	}
}

// Filter — an accessor: the callback was built elsewhere, this hands it out.
func (w *WebhookTriggerV2) Filter() func(identifier *model.ConsentEventV2UserIdentifier) bool {
	return w.filter
}

// scopes — a function returning no function type at all.
func (w *WebhookTriggerV2) Scopes() []string {
	return w.scopes
}

// --- Boundary class ---

// WithScope — the options convention (GID-126): a With-prefixed builder is
// named by its own convention.
func WithScope(scope string) model.SendOption {
	return func(identifier *model.ConsentEventV2UserIdentifier) {
		identifier.Value = scope
	}
}

// retryOption — the option TYPE alone exempts the builder: the suffix names
// the convention even without the With prefix.
func retryOption(value string) model.SendOption {
	return func(identifier *model.ConsentEventV2UserIdentifier) {
		identifier.Type = value
	}
}

// newest — boundary: the constructor prefix is a whole word, so a name merely
// starting with those letters is still judged.
func newest(scope string) func() string { // want `GID-273: function newest builds and returns a closure`
	return func() string { return scope }
}

// collect — the function literal inside belongs to the closure, not to this
// function: what is returned here is a slice.
func collect(scopes []string) []string {
	build := func(scope string) func() string {
		return func() string { return scope }
	}
	_ = build

	return scopes
}

// WithRetry — the With prefix alone exempts: the result is a bare function
// type, so the options suffix is not what keeps this quiet.
func WithRetry(count int) func(identifier *model.ConsentEventV2UserIdentifier) {
	return func(identifier *model.ConsentEventV2UserIdentifier) {
		identifier.Type = string(rune(count))
	}
}

// passthrough — boundary: this function DOES return a function type, but it
// builds nothing — the only literal here is nested inside another closure and
// belongs to it.
func passthrough(base func() string) func() string {
	wrap := func() func() string {
		return func() string { return base() }
	}
	_ = wrap

	return base
}
