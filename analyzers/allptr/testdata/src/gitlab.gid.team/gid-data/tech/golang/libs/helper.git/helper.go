// Minimal gdhelper stub for eval: AllPtr returns an iterator of
// pointers (range-over-func), just like the real library.
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
