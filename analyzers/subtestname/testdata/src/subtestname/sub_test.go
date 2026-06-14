// Eval for GID-191 (subtest names: no spaces or slashes).
package subtestname

import "testing"

const nameWithSpace = "has space"

// --- Positive cases (the violation is caught) ---

func TestPositive(t *testing.T) {
	t.Run("with space", func(t *testing.T) {}) // want `GID-191: subtest name "with space" contains a space\. Fix: use snake_case, go test -run 'Test/name' will not match it`
	t.Run("a/b", func(t *testing.T) {})        // want `GID-191: subtest name "a/b" contains a slash '/'\. Fix: use snake_case, go test -run 'Test/name' will not match it`
	t.Run(nameWithSpace, func(t *testing.T) {}) // want `GID-191: subtest name "has space" contains a space\. Fix: use snake_case, go test -run 'Test/name' will not match it`
}

func BenchmarkPositive(b *testing.B) {
	b.Run("x y", func(b *testing.B) {}) // want `GID-191: subtest name "x y" contains a space\. Fix: use snake_case, go test -run 'Test/name' will not match it`
}

// --- Negative cases (clean code passes) ---

func TestNegative(t *testing.T) {
	t.Run("ok_name", func(t *testing.T) {})
	t.Run("CamelCase", func(t *testing.T) {})
}

// --- Boundary cases ---

// table-driven: a name from tt.name is not a constant, not matched.
func TestTableDriven(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "with space"},
		{name: "a/b"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {})
	}
}

// A custom type with a Run(string, func) method — not *testing.T/*testing.B, not matched.
type fakeT struct{}

func (fakeT) Run(name string, fn func()) {}

func TestForeignRun(t *testing.T) {
	var ft fakeT
	ft.Run("with space", func() {})
	ft.Run("a/b", func() {})
}
