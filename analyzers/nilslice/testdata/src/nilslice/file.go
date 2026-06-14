// Eval for GID-185 (nil is a valid slice).
package nilslice

// --- Class 1: positive ---

// return with an empty slice literal.
func retEmptyInt() []int {
	return []int{} // want `GID-185: return nil instead of an empty slice\. Fix: a nil slice is valid`
}

// initialization via := with an empty literal.
func defineEmpty() {
	s := []string{} // want `GID-185: declare a zero-value slice\. Fix: var s \[\]T`
	_ = s
}

// initialization via var = with an empty literal.
var pkgEmpty = []byte{} // want `GID-185: declare a zero-value slice\. Fix: var s \[\]T`

func varEmptyLocal() {
	var s = []float64{} // want `GID-185: declare a zero-value slice\. Fix: var s \[\]T`
	_ = s
}

// --- Class 2: negative ---

// a nil slice in return — correct.
func retNil() []int {
	return nil
}

// a zero-value declaration — correct.
func varZero() {
	var s []int
	_ = s
}

// a non-empty literal is data, not "emptiness".
func retNonEmpty() []int {
	return []int{1, 2, 3}
}

func defineNonEmpty() {
	s := []string{"a"}
	_ = s
}

// --- Class 3: boundary (not matched) ---

func consume(_ []int) {}

type holder struct {
	X []int
}

// []T{} as a call argument — emptiness may be semantics (json [] vs null).
func emptyAsArg() {
	consume([]int{})
}

// []T{} as a struct field value — not matched.
func emptyAsField() {
	_ = holder{X: []int{}}
}

// an array [0]T{} is not a slice.
func emptyArray() {
	a := [0]int{}
	_ = a
}

// a map literal is not a slice.
func emptyMap() {
	m := map[string]int{}
	_ = m
}
