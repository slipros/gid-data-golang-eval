package deepequal

import "reflect"

// Позитив: reflect.DeepEqual запрещён.
func bad(a, b []int) bool {
	return reflect.DeepEqual(a, b) // want `GID-008: avoid reflect\.DeepEqual\. Fix: use require/cmp in tests or explicit field comparison in code\.`
}

// Негатив: другой вызов из reflect — не DeepEqual.
func good(a any) reflect.Kind {
	return reflect.TypeOf(a).Kind()
}
