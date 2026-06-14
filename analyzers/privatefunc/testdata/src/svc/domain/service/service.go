// Eval for GID-133 (privatefunc).
package service

import "context"

type Snapshot struct {
	name string
}

// A constructor also defines ownership: normalize below is used by the
// Snapshot (via the constructor) and Job entities — a shared helper, the norm.
func NewSnapshot() *Snapshot {
	return &Snapshot{name: normalize("snapshot")}
}

type Job struct{}

// --- Positive: used by only one entity ---

func (s *Snapshot) Render(ctx context.Context) string {
	return decorate(s.name) // the only consumer is Snapshot
}

func decorate(s string) string { // want `GID-133: private function "decorate" is used only by entity "Snapshot"\. Fix: make it a method`
	return ">" + s
}

// --- Positive: not used by anyone ---

func orphan() string { // want `GID-133: private function "orphan" belongs to the package\. Fix: make it a struct method`
	return ""
}

// --- Negative: a shared helper of two entities ---

func (j *Job) Name() string { return normalize("job") }

func normalize(s string) string {
	return s
}

// --- Not applicable: exported functions and init are not checked ---

func Shared(s string) string { return s }

func init() {}
