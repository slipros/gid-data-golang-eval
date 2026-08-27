// Vocabulary types for the GID-273 eval.
package model

// ConsentEventV2UserIdentifier — an identifier of the incident shape.
type ConsentEventV2UserIdentifier struct {
	Type  string
	Value string
}

// WebhookTriggerV2ContactFilterScope — the scope the filter is built for.
type WebhookTriggerV2ContactFilterScope string

// Predicate — a named function type: a filter of identifiers.
type Predicate func(identifier *ConsentEventV2UserIdentifier) bool

// SendOption — a named function type of the options convention.
type SendOption func(*ConsentEventV2UserIdentifier)
