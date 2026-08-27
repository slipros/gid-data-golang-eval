// Eval for GID-125, non-applicability: the module has no database — its
// /dal/repository speaks gRPC and fills these entities from protobuf, so a db
// tag would document a mapping that never reaches a column. The incident is
// consent-webhook-trigger (2026-08-27): 28 diagnostics on exactly this shape.
package entity

// Document — untagged on purpose: without a SQL stack in the module the rule
// stays silent.
type Document struct {
	ID     string
	Status string
	Owner  string
}

// Webhook — the same shape, filled from a protobuf message.
type Webhook struct {
	URL     string
	Attempt int
}
