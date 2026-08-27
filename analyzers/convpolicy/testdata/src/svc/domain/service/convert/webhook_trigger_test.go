// Non-applicability: a _test.go file is not judged (GID-250) — a test double
// reproducing the incident shape (enum constants picked by a length check) is
// scaffolding, not a converter deciding policy.
package convert

import "testing"

func TestEmptyReasonFromEvent(t *testing.T) {
	event := Event{}
	reason := WebhookTriggerV2EmptyReasonContactsFiltered
	if len(event.UserIdentifiers) == 0 {
		reason = WebhookTriggerV2EmptyReasonNoIdentifiers
	}
	if reason == "" {
		t.Fatal("empty reason")
	}
}
