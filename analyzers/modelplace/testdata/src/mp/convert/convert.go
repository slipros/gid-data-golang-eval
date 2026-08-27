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

// --- Part C: no function of the package hands out a struct declared here ---

// buildKind — positive: the unexported name kept the type out of part A, but
// the converter hands it outside anyway — the bypass part C closes.
func buildKind(code string) kind { // want `GID-270: function "buildKind" returns "kind" — a struct declared in this package, and a convert package holds no data types of its own\. Fix: declare the returned type in /domain/model \(or //nolint:gidmodelplace when explicitly intended\)`
	return kind{code: code}
}

// kindsByCode — positive: a container hands the struct out just the same.
func kindsByCode(codes []string) map[string]kind { // want `GID-270: function "kindsByCode" returns "kind"`
	return nil
}

// BuildFromRow — boundary: "Build" is already reported at its declaration by
// part A; the return does not add a second diagnostic.
func BuildFromRow(id string) *Build {
	return &Build{WebhookID: id}
}

// KindFromCode — negative: a named basic type is not a data struct.
func KindFromCode(code string) Kind {
	return Kind(code)
}

// MapperFor — negative: an interface is not a data struct either.
func MapperFor() Mapper {
	return nil
}
