// Eval GID-212: the signature check does not apply to _test.go files.
// Test functions and test helpers are not build functions — flagging them made
// it impossible to bring a build package with tests down to zero diagnostics.
package build

import "testing"

// A test function: signature (t *testing.T) with no results — not flagged.
func TestSelectJobs(t *testing.T) {
	if sql, _, _ := SelectJobs("new"); sql == "" {
		t.Fatal("empty sql")
	}
}

// A benchmark — not flagged.
func BenchmarkSelectJobs(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _, _ = SelectJobs("new")
	}
}

// A fuzz target — not flagged.
func FuzzSelectJobs(f *testing.F) {
	f.Fuzz(func(t *testing.T, status string) {
		_, _, _ = SelectJobs(status)
	})
}

// An exported test helper with an arbitrary signature — not flagged either:
// the whole file is out of the rule's scope.
func NewStatusFixture() []string {
	return []string{"new", "done"}
}
