// Eval GID-105: go-test entry points in a convert package are not converters —
// their names are mandated by `go test`, so they must NOT be flagged.
package convert

import "testing"

// --- Class 2: negative — test entry points, no diagnostics ---

func TestModelStatusFromEntity(t *testing.T) { _ = t }

func BenchmarkConvert(b *testing.B) { _ = b }

func Example() {}

func FuzzConvert(f *testing.F) { _ = f }

// --- Class 3: edge — an exported func merely starting with "Test" but with a
// lowercase next letter is NOT a test entry point; it is still judged (and here
// it happens to be a valid converter, so no diagnostic). ---

type testModel struct{ Name string }

type testEntity struct{ Name string }

func TestimonialFromModel(in testModel) testEntity { return testEntity{Name: in.Name} }
