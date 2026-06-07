// Eval для GID-006 (yoda-conditions).
package yodaconditions

const limit = 10

// --- Позитивные кейсы: нарушение ловится ---

func badStringLit(s string) bool {
	return "foo" == s // want `GID-006: переменная слева, литерал справа в сравнении`
}

func badNeq(n int) bool {
	return 0 != n // want `GID-006: переменная слева, литерал справа в сравнении`
}

// Граничный кейс: именованная константа слева — тоже йода.
func badConstName(n int) bool {
	return limit == n // want `GID-006: переменная слева, литерал справа в сравнении`
}

// --- Негативные кейсы: чистый код проходит ---

func goodOrder(s string) bool {
	return s == "foo"
}

func goodNeq(n int) bool {
	return n != 0
}

// --- Граничный кейс: const == const не матчится ---

func boundaryConstConst() bool {
	return limit == 10
}

// --- Неприменимость: сравнение двух переменных ---

func notApplicable(a, b int) bool {
	return a == b
}
