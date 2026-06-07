// Класс «неприменимость»: обычный .go файл без testing —
// здесь нет t.Run/b.Run, диагностик быть не должно.
package plain

type Runner struct{}

// Run с похожей сигнатурой, но не из пакета testing.
func (Runner) Run(name string, fn func()) {}

func use() {
	var r Runner
	r.Run("with space", func() {})
	r.Run("a/b", func() {})
}
