package yoda

// Позитив: литерал слева — yoda-условие.
func bad(x int) bool {
	return 0 == x // want `GID-006: yoda condition — the literal must be on the right\. Fix: write "x == 0" instead of "0 == x"\.`
}

// Позитив: то же для !=.
func bad2(x int) bool {
	return 5 != x // want `GID-006: yoda condition`
}

// Негатив: переменная слева, литерал справа.
func good(x int) bool {
	return x == 0
}

// Неприменимость: обе стороны не-константы.
func boundary(a, b int) bool {
	return a == b
}

// Неприменимость: обе стороны константы.
func boundaryConst() bool {
	return 1 == 1
}
