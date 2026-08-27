// Eval for GID-270 (model-place): a _test.go file is not judged (GID-250) —
// a struct a test declares is its own scaffolding.
package service

// TestHarness — an exported struct declared by a test: outside the rule.
type TestHarness struct {
	svc *Processor
}

func (h *TestHarness) boot() {}
