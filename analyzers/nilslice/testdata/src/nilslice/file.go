// Eval для GID-185 (nil is a valid slice).
package nilslice

// --- Класс 1: позитивные ---

// return пустым литералом слайса.
func retEmptyInt() []int {
	return []int{} // want `GID-185: возвращайте nil вместо пустого слайса — nil-слайс валиден`
}

// инициализация через := пустым литералом.
func defineEmpty() {
	s := []string{} // want `GID-185: объявляйте zero-value слайс: var s \[\]T`
	_ = s
}

// инициализация через var = пустым литералом.
var pkgEmpty = []byte{} // want `GID-185: объявляйте zero-value слайс: var s \[\]T`

func varEmptyLocal() {
	var s = []float64{} // want `GID-185: объявляйте zero-value слайс: var s \[\]T`
	_ = s
}

// --- Класс 2: негативные ---

// nil-слайс в return — правильно.
func retNil() []int {
	return nil
}

// zero-value объявление — правильно.
func varZero() {
	var s []int
	_ = s
}

// непустой литерал — это данные, не «пустота».
func retNonEmpty() []int {
	return []int{1, 2, 3}
}

func defineNonEmpty() {
	s := []string{"a"}
	_ = s
}

// --- Класс 3: граничные (не матчатся) ---

func consume(_ []int) {}

type holder struct {
	X []int
}

// []T{} аргументом вызова — пустота может быть семантикой (json [] vs null).
func emptyAsArg() {
	consume([]int{})
}

// []T{} значением поля структуры — не матчим.
func emptyAsField() {
	_ = holder{X: []int{}}
}

// массив [0]T{} — не слайс.
func emptyArray() {
	a := [0]int{}
	_ = a
}

// map-литерал — не слайс.
func emptyMap() {
	m := map[string]int{}
	_ = m
}
