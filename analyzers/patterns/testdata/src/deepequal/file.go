package deepequal

import "reflect"

// Positive: reflect.DeepEqual is forbidden.
func bad(a, b []int) bool {
	return reflect.DeepEqual(a, b) // want `GID-008: avoid reflect\.DeepEqual\. Fix: use require/cmp in tests or explicit field comparison in code\.`
}

// Negative: another reflect call — not DeepEqual.
func good(a any) reflect.Kind {
	return reflect.TypeOf(a).Kind()
}
