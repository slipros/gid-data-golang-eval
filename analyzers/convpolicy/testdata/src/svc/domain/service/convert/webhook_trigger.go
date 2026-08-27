package convert

// Event is the input of the webhook-trigger converter.
type Event struct {
	UserIdentifiers Identifiers
}

// Identifiers is a slice with a domain filter, as the webhook event carries.
type Identifiers []string

// Filter returns the identifiers that pass the whitelist.
func (ids Identifiers) Filter() Identifiers {
	return ids
}

// WebhookTriggerV2EmptyReason is the named enum the converter picks among.
type WebhookTriggerV2EmptyReason string

const (
	WebhookTriggerV2EmptyReasonContactsFiltered WebhookTriggerV2EmptyReason = "contacts_filtered"
	WebhookTriggerV2EmptyReasonNoIdentifiers    WebhookTriggerV2EmptyReason = "no_identifiers"
	WebhookTriggerV2EmptyReasonWhitelist        WebhookTriggerV2EmptyReason = "whitelist"
)

// Kind is the input enum of the mapping converters below.
type Kind string

const (
	KindFiltered Kind = "filtered"
	KindNoIDs    Kind = "no_ids"
	KindNone     Kind = "none"
)

// --- positive: enum constants picked by a non-enum condition (the
// consent-webhook-trigger incident shape) ---

// EmptyReasonFromEvent decides which reason to show outside by lengths of the
// input — business policy in a converter, not enum→enum mapping.
func EmptyReasonFromEvent(event Event, source []string) WebhookTriggerV2EmptyReason {
	reason := WebhookTriggerV2EmptyReasonContactsFiltered
	if len(event.UserIdentifiers) == 0 {
		reason = WebhookTriggerV2EmptyReasonNoIdentifiers // want `GID-247: convert function "EmptyReasonFromEvent" branches on a non-enum condition to select among enum constants of "WebhookTriggerV2EmptyReason" for "reason" — the converter decides a business question, not mapping. Fix: move the decision to /domain/model \(a clean predicate or a factory\) and pass the converter the ready result`
	}
	identifiers := event.UserIdentifiers.Filter()
	if len(identifiers) == 0 && len(source) > 0 {
		reason = WebhookTriggerV2EmptyReasonWhitelist
	}

	return reason
}

// FlagPickedReason picks between enum constants by a bool flag of the input.
func FlagPickedReason(event Event, whitelisted bool) WebhookTriggerV2EmptyReason {
	reason := WebhookTriggerV2EmptyReasonNoIdentifiers
	if whitelisted {
		reason = WebhookTriggerV2EmptyReasonWhitelist // want `GID-247`
	}

	return reason
}

// --- negative: enum→enum mapping — the condition reads only enum values ---

// ReasonFromKindIf maps one enum to another through if/else on the enum value
// — vocabulary translation (GID-143/233 territory), not policy.
func ReasonFromKindIf(kind Kind) WebhookTriggerV2EmptyReason {
	reason := WebhookTriggerV2EmptyReasonNoIdentifiers
	if kind == KindFiltered {
		reason = WebhookTriggerV2EmptyReasonContactsFiltered
	} else if kind == KindNone {
		reason = WebhookTriggerV2EmptyReasonWhitelist
	}

	return reason
}

// ReasonFromKindSwitch maps one enum to another through a switch on the enum
// value — silent, same as CodecFromSource.
func ReasonFromKindSwitch(kind Kind) WebhookTriggerV2EmptyReason {
	var reason WebhookTriggerV2EmptyReason
	switch kind {
	case KindFiltered:
		reason = WebhookTriggerV2EmptyReasonContactsFiltered
	case KindNoIDs:
		reason = WebhookTriggerV2EmptyReasonNoIdentifiers
	default:
		reason = WebhookTriggerV2EmptyReasonWhitelist
	}

	return reason
}

// --- boundary: mixed condition — reads an enum AND a non-enum value ---

// ReasonFromKindAndSize branches on a mixed condition: the enum comparison
// alone would be a mapping, but the length check makes the branch decide by
// input — still policy.
func ReasonFromKindAndSize(kind Kind, ids []string) WebhookTriggerV2EmptyReason {
	reason := WebhookTriggerV2EmptyReasonContactsFiltered
	if kind == KindFiltered && len(ids) == 0 {
		reason = WebhookTriggerV2EmptyReasonNoIdentifiers // want `GID-247`
	}

	return reason
}

// --- boundary: the same enum constant in every branch (no selection) ---

// SameReason assigns one distinct enum constant regardless of the input — not
// a selection.
func SameReason(event Event) WebhookTriggerV2EmptyReason {
	reason := WebhookTriggerV2EmptyReasonContactsFiltered
	if len(event.UserIdentifiers) == 0 {
		reason = WebhookTriggerV2EmptyReasonContactsFiltered
	}

	return reason
}
