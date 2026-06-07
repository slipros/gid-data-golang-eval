package newderef

type point struct{ X, Y int }

// Позитив: new() запрещён.
func bad() *point {
	return new(point) // want `GID-005: avoid the new\(\) builtin\. Fix: use "&T\{\}" for structs or "var x T" instead of "new\(T\)"\.`
}

// Негатив: &T{} вместо new(T).
func good() *point {
	return &point{}
}

// Неприменимость: пользовательская функция с именем new — не builtin.
func boundary() int {
	newx := func(v int) int { return v }
	return newx(1)
}
