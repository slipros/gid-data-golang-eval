// Eval of GID-272 in the application-module layout: pkg/<module>/domain/**
// is judged the same as internal/domain/**. The fixture is the incident shape
// (2026-08-27, consent-webhook-trigger): six arguments, four of them maps
// keyed by the same organizationID — the four maps are one thing and ask to be
// grouped into a type.
package convert

func WebhooksTriggersV2FromConsentEventV2( // want `GID-272: function WebhooksTriggersV2FromConsentEventV2 takes 6 substantive arguments \(allowed: 3\)`
	organizationID string,
	triggers map[string][]Trigger,
	disabled map[string]bool,
	events map[string][]Event,
	fallback map[string]string,
	limit int,
) []TriggerV2 {
	return nil
}

// TriggerV2 is the output shape.
type TriggerV2 struct {
	ID string
}

// Trigger is an event trigger.
type Trigger struct {
	ID string
}

// Event is an event payload.
type Event struct {
	ID string
}
