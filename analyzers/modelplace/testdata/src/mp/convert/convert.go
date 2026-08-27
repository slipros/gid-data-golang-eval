// Eval for GID-270 (model-place), part A: a convert package declares no types.
package convert

// Build — positive: an exported struct declared in a convert package. The
// incident shape is consent-webhook-trigger's WebhookTriggerV2Build
// (internal/domain/usecase/convert/webhook_trigger_v2.go, 2026-08-27).
type Build struct { // want `GID-270: type "Build" is declared in a convert package — a converter transforms foreign types and has no type vocabulary of its own\. Fix: declare the type in /domain/model \(or //nolint:gidmodelplace when explicitly intended\)`
	WebhookID string
	Payload   []byte
	Attempt   int
}

// kind — non-applicability: an unexported type is a package detail that does
// not leak.
type kind struct {
	code string
}

// Kind — boundary: a named basic type in convert is not a struct declaration.
type Kind string

// Mapper — boundary: an interface in convert is not judged either, only
// struct declarations are.
type Mapper interface {
	Map(string) string
}
