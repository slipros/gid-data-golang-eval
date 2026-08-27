// Eval for GID-273: a _test.go file is not judged (GID-250) — a helper
// returning a cleanup func is named by the testing convention.
package usecase

func setup(scope string) func() {
	return func() { _ = scope }
}
