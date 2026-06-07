// Eval для GID-008 (no-deepequal).
package nodeepequal

import "reflect"

// --- Позитивные кейсы: нарушение ловится ---

func badEqual(a, b []int) bool {
	return reflect.DeepEqual(a, b) // want `GID-008: в тестах — cmp/require, в коде — явное сравнение вместо reflect\.DeepEqual`
}

// Граничный кейс: DeepEqual со структурами.
type point struct {
	x, y int
}

func badStruct(a, b point) bool {
	return reflect.DeepEqual(a, b) // want `GID-008: в тестах — cmp/require, в коде — явное сравнение вместо reflect\.DeepEqual`
}

// --- Негативные кейсы: явное сравнение проходит ---

func goodCompare(a, b point) bool {
	return a == b
}

// --- Неприменимость: другие функции reflect не трогаем ---

func notApplicable(a int) reflect.Kind {
	return reflect.TypeOf(a).Kind()
}
