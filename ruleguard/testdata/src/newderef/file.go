// Eval для GID-005 (new-deref).
package newderef

type job struct {
	id int
}

// --- Позитивные кейсы: нарушение ловится ---

func badStruct() *job {
	return new(job) // want `GID-005: используйте &T\{\} для структур или var x T вместо new\(job\)`
}

// Граничный кейс: new() для примитива тоже матчим (стайлгайд предпочитает var).
func badPrimitive() *int {
	return new(int) // want `GID-005: используйте &T\{\} для структур или var x T вместо new\(int\)`
}

// --- Негативные кейсы: чистый код проходит ---

func goodStruct() *job {
	return &job{}
}

func goodVar() job {
	var j job
	return j
}

// --- Неприменимость: make для слайсов/мап — не new ---

func notApplicable() []int {
	return make([]int, 0)
}
