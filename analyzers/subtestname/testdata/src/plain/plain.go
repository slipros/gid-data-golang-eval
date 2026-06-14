// The "not applicable" class: an ordinary .go file without testing —
// there is no t.Run/b.Run here, there must be no diagnostics.
package plain

type Runner struct{}

// Run with a similar signature, but not from the testing package.
func (Runner) Run(name string, fn func()) {}

func use() {
	var r Runner
	r.Run("with space", func() {})
	r.Run("a/b", func() {})
}
