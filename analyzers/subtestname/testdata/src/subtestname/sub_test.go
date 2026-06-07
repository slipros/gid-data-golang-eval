// Eval для GID-191 (subtest names: без пробелов и слешей).
package subtestname

import "testing"

const nameWithSpace = "has space"

// --- Позитивные кейсы (нарушение ловится) ---

func TestPositive(t *testing.T) {
	t.Run("with space", func(t *testing.T) {}) // want `GID-191: имя subtest "with space" содержит пробел — используйте snake_case: go test -run 'Test/имя' не найдёт его`
	t.Run("a/b", func(t *testing.T) {})        // want `GID-191: имя subtest "a/b" содержит слеш '/' — используйте snake_case: go test -run 'Test/имя' не найдёт его`
	t.Run(nameWithSpace, func(t *testing.T) {}) // want `GID-191: имя subtest "has space" содержит пробел — используйте snake_case: go test -run 'Test/имя' не найдёт его`
}

func BenchmarkPositive(b *testing.B) {
	b.Run("x y", func(b *testing.B) {}) // want `GID-191: имя subtest "x y" содержит пробел — используйте snake_case: go test -run 'Test/имя' не найдёт его`
}

// --- Негативные кейсы (чистый код проходит) ---

func TestNegative(t *testing.T) {
	t.Run("ok_name", func(t *testing.T) {})
	t.Run("CamelCase", func(t *testing.T) {})
}

// --- Граничные кейсы ---

// table-driven: имя из tt.name — не константа, не матчится.
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

// Свой тип с методом Run(string, func) — не *testing.T/*testing.B, не матчится.
type fakeT struct{}

func (fakeT) Run(name string, fn func()) {}

func TestForeignRun(t *testing.T) {
	var ft fakeT
	ft.Run("with space", func() {})
	ft.Run("a/b", func() {})
}
