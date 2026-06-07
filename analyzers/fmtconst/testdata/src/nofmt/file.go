// Eval GID-186, класс «неприменимость»: пакет без fmt/log/pkg-errors —
// printf-функций нет, диагностики быть не должно.
package nofmt

func describe(s string) string {
	return s + "!"
}

func combine(a, b string) string {
	return a + b
}
