// Eval of GID-186, "inapplicability" class: a package without fmt/log/pkg-errors —
// no printf functions, there must be no diagnostic.
package nofmt

func describe(s string) string {
	return s + "!"
}

func combine(a, b string) string {
	return a + b
}
