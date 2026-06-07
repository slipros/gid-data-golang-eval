// Минимальный stub gdhelper для eval: AllPtr возвращает итератор
// указателей (range-over-func), как и реальная библиотека.
package gdhelper

func AllPtr[S ~[]E, E any](s S) func(yield func(int, *E) bool) {
	return func(yield func(int, *E) bool) {
		for i := range s {
			if !yield(i, &s[i]) {
				return
			}
		}
	}
}
